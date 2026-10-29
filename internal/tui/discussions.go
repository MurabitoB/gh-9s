package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/MurabitoB/gh-9s/internal/ghcli"
	"github.com/MurabitoB/gh-9s/internal/ui"
)

// Outside calls the discussion screens make; tests replace them.
var (
	fetchDiscList = func(tg ghcli.RepoRef, f ghcli.DiscFilter, limit int) (ghcli.DiscussionList, error) {
		return ghcli.FetchDiscussions(hostOrDefault(tg.Host), tg, f, limit)
	}
	fetchDisc = func(tg ghcli.RepoRef, number string) (ghcli.DiscussionDetail, error) {
		return ghcli.FetchDiscussion(hostOrDefault(tg.Host), tg, number)
	}
	commentDisc = func(tg ghcli.RepoRef, discussionID, body, replyTo string) error {
		return ghcli.AddDiscussionComment(hostOrDefault(tg.Host), discussionID, body, replyTo)
	}
	closeDisc = func(tg ghcli.RepoRef, discussionID, reason string) error {
		return ghcli.CloseDiscussion(hostOrDefault(tg.Host), discussionID, reason)
	}
	reopenDisc = func(tg ghcli.RepoRef, discussionID string) error {
		return ghcli.ReopenDiscussion(hostOrDefault(tg.Host), discussionID)
	}
	markDiscAnswer = func(tg ghcli.RepoRef, commentID string, mark bool) error {
		return ghcli.MarkDiscussionAnswer(hostOrDefault(tg.Host), commentID, mark)
	}
	lockDisc = func(tg ghcli.RepoRef, discussionID string, lock bool) error {
		return ghcli.LockDiscussion(hostOrDefault(tg.Host), discussionID, lock)
	}
)

// discFilterOf is the filter the list reads with: the state and answer filters and the category.
func (m model) discFilterOf() ghcli.DiscFilter {
	f := m.discFilter
	f.Category = m.discCat
	return f
}

// discListLimit is how many discussions the list reads.
func (m model) discListLimit() int {
	return max(m.discLimit, ghcli.DiscussionListLimit)
}

// discFilterName words the state and answer filters for the title ("" for the plain open list).
func discFilterName(f ghcli.DiscFilter) string {
	var parts []string
	switch f.State {
	case "closed", "all":
		parts = append(parts, f.State)
	}
	switch f.Answer {
	case "answered", "unanswered":
		parts = append(parts, f.Answer)
	}
	return strings.Join(parts, " + ")
}

// discListState is what the list screen keeps: what was read, for the categories dialog and the pane.
type discListState struct{ l ghcli.DiscussionList }

// discState is what the discussion screen keeps. The selected comment is kept by id, so a refresh keeps it.
type discState struct {
	d      ghcli.DiscussionDetail
	scroll int
	sel    string
}

func init() {
	register(viewDiscussions, screen{
		name:     "discussions",
		identity: func(m model) string { return m.discListID() },
		fetch: func(m model) tea.Cmd {
			tg, f, limit, id := m.curTarget(), m.discFilterOf(), m.discListLimit(), m.discListID()
			return func() tea.Msg {
				l, err := fetchDiscList(tg, f, limit)
				return screenMsg{view: viewDiscussions, id: id, data: l, err: err}
			}
		},
		apply: func(m model, msg screenMsg) model {
			l := msg.data.(ghcli.DiscussionList)
			m = m.withScreenData(viewDiscussions, discListState{l})
			m.tbl = discTable(l)
			if n := len(m.visible()); m.cursor >= n {
				m.cursor = max(0, n-1)
			}
			if l.Enabled && l.Total > len(l.Rows) {
				m.status = styDim.Render(fmt.Sprintf("showing the first %d of %d discussions (m loads more, c narrows them by category)", len(l.Rows), l.Total))
			}
			return m
		},
		key:    discListKey,
		body:   discListBody,
		footer: "enter open  O/C/A state  u unanswered  a answered  m more  c category  o browser",
	})
	register(viewDiscussion, screen{
		name:     "discussion",
		identity: func(m model) string { return m.itemID(m.disc) + "|discussion" },
		fetch: func(m model) tea.Cmd {
			tg, number, id := m.curTarget(), m.disc, m.itemID(m.disc)+"|discussion"
			return func() tea.Msg {
				d, err := fetchDisc(tg, number)
				return screenMsg{view: viewDiscussion, id: id, data: d, err: err}
			}
		},
		apply: func(m model, msg screenMsg) model {
			old, _ := m.screenData(viewDiscussion).(discState)
			return m.withScreenData(viewDiscussion, discState{d: msg.data.(ghcli.DiscussionDetail), scroll: old.scroll, sel: old.sel})
		},
		key:    discKey,
		body:   discBody,
		footer: "C comment  r reply  n/N pick  x/X close/reopen  A answer  L lock  o open  j/k scroll",
	})
	dialogHandlers["disc.category"] = func(m model, r dialogResultMsg) (model, tea.Cmd) {
		st, _ := m.screenData(viewDiscussions).(discListState)
		if !r.ok || len(r.values) != 1 || m.view != viewDiscussions {
			return m, nil
		}
		id, name := "", ""
		if r.values[0] != discAllCategories {
			for _, c := range st.l.Categories {
				if c.ID == r.values[0] {
					id, name = c.ID, c.Name
				}
			}
			if id == "" {
				return m, nil // a category the list no longer has
			}
		}
		if id == m.discCat {
			return m, nil
		}
		m.discCat, m.discCatName, m.discLimit = id, name, 0
		m.cursor, m.top, m.loading = 0, 0, true
		m.tbl = table{}
		return m, m.fetch()
	}
	dialogHandlers["disc.close"] = func(m model, r dialogResultMsg) (model, tea.Cmd) {
		p := m.pending
		m.pending = nil
		if !r.ok || p == nil || len(r.values) != 1 {
			return m, nil
		}
		return m, func() tea.Msg {
			return prActionDoneMsg{text: "closed " + p.label, err: closeDisc(p.target, p.discID, r.values[0])}
		}
	}
	discConfirmed := func(text string, do func(p *prAction) error) func(model, dialogResultMsg) (model, tea.Cmd) {
		return func(m model, r dialogResultMsg) (model, tea.Cmd) {
			p := m.pending
			m.pending = nil
			if !r.ok || p == nil {
				return m, nil
			}
			return m, func() tea.Msg { return prActionDoneMsg{text: text + " " + p.label, err: do(p)} }
		}
	}
	dialogHandlers["disc.mark"] = discConfirmed("marked the answer of", func(p *prAction) error { return markDiscAnswer(p.target, p.commentID, true) })
	dialogHandlers["disc.unmark"] = discConfirmed("removed the answer mark from", func(p *prAction) error { return markDiscAnswer(p.target, p.commentID, false) })
	dialogHandlers["disc.lock"] = discConfirmed("locked", func(p *prAction) error { return lockDisc(p.target, p.discID, true) })
	dialogHandlers["disc.unlock"] = discConfirmed("unlocked", func(p *prAction) error { return lockDisc(p.target, p.discID, false) })
	dialogHandlers["disc.comment"] = func(m model, r dialogResultMsg) (model, tea.Cmd) {
		p := m.pending
		m.pending = nil
		if !r.ok || p == nil {
			return m, nil
		}
		verb := "commented on "
		if p.replyTo != "" {
			verb = "replied in "
		}
		m.discBusy = true // cleared by the answer, whatever it is
		return m, func() tea.Msg {
			return prActionDoneMsg{text: verb + p.label, err: commentDisc(p.target, p.discID, r.text, p.replyTo)}
		}
	}
}

const discAllCategories = "\x00all" // the value of "All categories": no real id looks like this

// discTable lays the list out.
func discTable(l ghcli.DiscussionList) table {
	t := table{cols: []string{"#", "CATEGORY", "TITLE", "AUTHOR", "COMMENTS", "UPDATED"}}
	for _, d := range l.Rows {
		st, title := stOK, d.Title
		if d.Closed {
			st = stOther
		}
		if d.Locked {
			title = "[locked] " + title
		}
		if d.Closed {
			title = "[closed] " + title
		}
		if d.Answered {
			title = "✓ " + title
		}
		t.rows = append(t.rows, row{id: strconv.Itoa(d.Number), state: st, url: d.URL, updated: d.UpdatedAt,
			cells: []string{strconv.Itoa(d.Number), orDash(d.Category), title, d.Author, strconv.Itoa(d.Comments), age(d.UpdatedAt)}})
	}
	return t
}

func discListBody(m model) string {
	if st, ok := m.screenData(viewDiscussions).(discListState); ok && !m.loading && m.err == nil && !st.l.Enabled && m.curTarget().Name != "" {
		return " Discussions are not enabled in " + displayRef(m.curTarget()) + "\n"
	}
	return m.tableBody()
}

// discListID names what the list shows, so a late answer for another filter or category is dropped.
func (m model) discListID() string {
	f := m.discFilterOf()
	return fmt.Sprintf("%s|%s|%s|%s|%d", m.curTarget(), f.Category, f.State, f.Answer, m.discListLimit())
}

// discListKey: O/C/A pick the state, u / a the answered filter, m reads more, c picks the category, enter opens a
// discussion; the rest is the shared table behaviour.
func discListKey(m model, k tea.KeyMsg) (model, tea.Cmd, bool) {
	refetch := func(m model) (model, tea.Cmd, bool) {
		m.cursor, m.top, m.loading = 0, 0, true
		return m, m.fetch(), true
	}
	switch k.String() {
	case "O", "C", "A":
		m.discFilter.State = map[string]string{"O": "", "C": "closed", "A": "all"}[k.String()]
		m.discLimit = 0
		return refetch(m)
	case "u", "a":
		want := map[string]string{"u": "unanswered", "a": "answered"}[k.String()]
		if m.discFilter.Answer == want {
			want = ""
		}
		m.discFilter.Answer, m.discLimit = want, 0
		return refetch(m)
	case "m":
		st, _ := m.screenData(viewDiscussions).(discListState)
		if !st.l.More {
			m.status = styDim.Render("no more discussions")
			return m, nil, true
		}
		if m.discListLimit() >= ghcli.DiscussionListMax {
			m.status = styDim.Render(fmt.Sprintf("the list stops at %d discussions: c, O/C/A and u/a narrow it", ghcli.DiscussionListMax))
			return m, nil, true
		}
		m.discLimit = min(m.discListLimit()+ghcli.DiscussionListLimit, ghcli.DiscussionListMax)
		m.loading = true // keep the cursor and scroll: the same rows come first
		return m, m.fetch(), true
	case "c":
		st, _ := m.screenData(viewDiscussions).(discListState)
		if len(st.l.Categories) == 0 {
			m.status = styDim.Render("no categories to pick from (yet)")
			return m, nil, true
		}
		cs := []choice{{label: "All categories", value: discAllCategories, disabled: m.discCat == "", note: boolNote(m.discCat == "", "current")}}
		for _, c := range st.l.Categories {
			cs = append(cs, choice{label: c.Name, value: c.ID, disabled: c.ID == m.discCat, note: boolNote(c.ID == m.discCat, "current")})
		}
		return m.withDialog(newChoiceDialog("disc.category", "Show discussions of…", cs, false, nil)), nil, true
	case "enter":
		r, ok := m.selected()
		if !ok {
			return m, nil, true
		}
		m.disc = r.id
		m = m.withScreenData(viewDiscussion, discState{}) // a different discussion starts at the top
		cmd := m.push(viewDiscussion)
		return m, cmd, true
	}
	return m, nil, false
}

func boolNote(b bool, s string) string {
	if b {
		return s
	}
	return ""
}

// discPreviewLines is the pane for the selected discussion of the list: the list already read its text.
func (m model) discPreviewLines(width int) ([]string, bool) {
	if m.view != viewDiscussions {
		return nil, false
	}
	r, ok := m.selected()
	if !ok {
		return nil, false
	}
	st, _ := m.screenData(viewDiscussions).(discListState)
	for _, d := range st.l.Rows {
		if strconv.Itoa(d.Number) != r.id {
			continue
		}
		width = max(width, 10)
		var doc detailDoc
		addWrapped(&doc, fmt.Sprintf("#%d %s", d.Number, d.Title), width, bold)
		doc.add(ui.Dim.Render(discMeta(d.Author, d.Category, d.Answered, d.Closed, d.Locked)))
		description(&doc, d.Body, width)
		lines := doc.lines
		if len(lines) > previewTotalLines {
			lines = append(lines[:previewTotalLines:previewTotalLines], ui.Dim.Render("… more in the discussion (enter)"))
		}
		return lines, true
	}
	return nil, false
}

// discMeta words who wrote it, where, and what state it is in.
func discMeta(author, category string, answered, closed, locked bool) string {
	parts := []string{author}
	if category != "" {
		parts = append(parts, category)
	}
	if answered {
		parts = append(parts, "✓ answered")
	}
	if closed {
		parts = append(parts, "closed")
	}
	if locked {
		parts = append(parts, "locked")
	}
	return strings.Join(parts, " · ")
}

// ---- one discussion -----------------------------------------------------------------------------------------------

// discEntry is a comment or a reply on the screen: where its lines are, and which thread it belongs to.
type discEntry struct {
	id, parent, author string // parent is the top-level comment of a reply, "" for a comment
	start, end         int    // its lines, end exclusive
}

type discDoc struct {
	lines   []string
	entries []discEntry
}

// discLines lays the discussion out at a width.
func discLines(d ghcli.DiscussionDetail, width int) discDoc {
	var out discDoc
	if d.Number == 0 && d.Title == "" {
		return discDoc{lines: []string{"loading…"}}
	}
	out.lines = append(out.lines, fmt.Sprintf("#%d  %s", d.Number, d.Title), discMeta(d.Author, d.Category, d.Answered, d.Closed, d.Locked))
	var body detailDoc
	body.add("", "── Description ──")
	if strings.TrimSpace(d.Body) == "" {
		body.add("(no description)")
	} else {
		body.addMarkdown(d.Body, width, "")
	}
	out.lines = append(out.lines, body.lines...)
	if d.CommentTotal == 0 && len(d.Comments) == 0 {
		out.lines = append(out.lines, "", "── No comments yet ──")
		return out
	}
	out.lines = append(out.lines, "", fmt.Sprintf("── Comments (%d) ──", d.CommentTotal))
	for _, c := range d.Comments {
		start := len(out.lines)
		head := fmt.Sprintf("%s commented · %s", c.Author, ui.Age(c.CreatedAt))
		if c.IsAnswer {
			head += "   ✓ answer"
		}
		out.lines = append(out.lines, "", head)
		var cb detailDoc
		cb.addMarkdown(c.Body, width, "  ")
		out.lines = append(out.lines, cb.lines...)
		out.entries = append(out.entries, discEntry{id: c.ID, author: c.Author, start: start + 1, end: len(out.lines)})
		for _, r := range c.Replies {
			rs := len(out.lines)
			out.lines = append(out.lines, "", fmt.Sprintf("    %s replied · %s", r.Author, ui.Age(r.CreatedAt)))
			var rb detailDoc
			rb.addMarkdown(r.Body, width, "      ")
			out.lines = append(out.lines, rb.lines...)
			out.entries = append(out.entries, discEntry{id: r.ID, parent: c.ID, author: r.Author, start: rs + 1, end: len(out.lines)})
		}
		if c.ReplyTotal > len(c.Replies) {
			out.lines = append(out.lines, ui.Dim.Render(fmt.Sprintf("    … %d more replies not shown (o opens the discussion in the browser)", c.ReplyTotal-len(c.Replies))))
		}
	}
	if d.CommentTotal > len(d.Comments) {
		out.lines = append(out.lines, "", ui.Dim.Render(fmt.Sprintf("… %d more comments not shown (o opens the discussion in the browser)", d.CommentTotal-len(d.Comments))))
	}
	return out
}

func (doc discDoc) entry(id string) (discEntry, int, bool) {
	for i, e := range doc.entries {
		if e.id == id && id != "" {
			return e, i, true
		}
	}
	return discEntry{}, -1, false
}

func discBody(m model) string {
	st, _ := m.screenData(viewDiscussion).(discState)
	h := m.logHeight()
	var b strings.Builder
	if m.err != nil {
		b.WriteString(ui.Err.Render(" "+m.err.Error()) + "\n")
		h--
	}
	doc := discLines(st.d, m.contentWidth())
	sel, _, has := doc.entry(st.sel)
	n := 0
	for i := min(st.scroll, max(0, len(doc.lines)-1)); i < len(doc.lines) && n < h; i++ {
		margin := " "
		if has && i >= sel.start && i < sel.end {
			margin = ui.Key.Render("▌")
		}
		b.WriteString(margin + doc.lines[i] + "\n")
		n++
	}
	for ; n < h; n++ {
		b.WriteString("\n")
	}
	return b.String()
}

// discAction is a comment or reply being written: everything it needs is taken when the key is pressed.
func (m model) startDiscComment(d ghcli.DiscussionDetail, replyTo, who string) (model, tea.Cmd) {
	label := displayRef(m.curTarget()) + "#" + m.disc
	p := &prAction{target: m.curTarget(), number: m.disc, label: label, discID: d.ID, replyTo: replyTo}
	m.pending = p
	title := "Comment on " + label
	if replyTo != "" {
		title = "Reply to " + who + " in " + label
	}
	return m.withDialog(newTextDialog("disc.comment", title, "Your comment")), nil
}

func discKey(m model, k tea.KeyMsg) (model, tea.Cmd, bool) {
	st, _ := m.screenData(viewDiscussion).(discState)
	h := m.logHeight()
	doc := discLines(st.d, m.contentWidth())
	maxScroll := max(0, len(doc.lines)-h)
	switch k.String() {
	case "j", "down":
		st.scroll = min(st.scroll+1, maxScroll)
	case "k", "up":
		st.scroll = max(st.scroll-1, 0)
	case "ctrl+d", "pgdown", " ":
		st.scroll = min(st.scroll+h/2, maxScroll)
	case "ctrl+u", "pgup":
		st.scroll = max(st.scroll-h/2, 0)
	case "g", "home":
		st.scroll = 0
	case "G", "end":
		st.scroll = maxScroll
	case "n", "N":
		if len(doc.entries) == 0 {
			m.status = styDim.Render("no comments to select")
			return m, nil, true
		}
		_, i, has := doc.entry(st.sel)
		switch {
		case !has && k.String() == "n":
			i = 0
			for j, e := range doc.entries { // the first comment on screen
				if e.end > st.scroll {
					i = j
					break
				}
			}
		case !has:
			i = len(doc.entries) - 1
		case k.String() == "n":
			i = min(i+1, len(doc.entries)-1)
		default:
			i = max(i-1, 0)
		}
		e := doc.entries[i]
		st.sel = e.id
		if e.start < st.scroll {
			st.scroll = e.start
		} else if e.end > st.scroll+h {
			st.scroll = min(e.start, e.end-h)
		}
		st.scroll = min(max(st.scroll, 0), maxScroll)
	case "C", "r":
		if m.discBusy {
			m.status = styDim.Render("a comment is still being sent")
			return m, nil, true
		}
		if st.d.ID == "" {
			m.status = styDim.Render("still loading the discussion")
			return m, nil, true
		}
		if st.d.Locked {
			m.status = styDim.Render("this discussion is locked — nothing can be added")
			return m, nil, true
		}
		if k.String() == "C" {
			next, cmd := m.startDiscComment(st.d, "", "")
			return next, cmd, true
		}
		e, _, has := doc.entry(st.sel)
		if !has {
			m.status = styDim.Render("select a comment first (n / N)")
			return m, nil, true
		}
		thread := e.parent
		if thread == "" {
			thread = e.id
		}
		next, cmd := m.startDiscComment(st.d, thread, e.author)
		return next, cmd, true
	case "x", "X", "A", "L":
		next, cmd := m.discStateKey(k.String(), st, doc)
		return next, cmd, true
	case "o":
		if st.d.URL != "" {
			m = m.open(st.d.URL)
		}
		return m, nil, true
	default:
		return m, nil, false
	}
	return m.withScreenData(viewDiscussion, st), nil, true
}

// discStateKey handles the keys of the discussion that change it: x close, X reopen, A mark / unmark the selected
// comment as the answer, L lock / unlock. Each says why when it does not apply.
func (m model) discStateKey(key string, st discState, doc discDoc) (model, tea.Cmd) {
	note := func(text string) (model, tea.Cmd) {
		m.status = styDim.Render(text)
		return m, nil
	}
	d := st.d
	if d.ID == "" {
		return note("still loading the discussion")
	}
	label := displayRef(m.curTarget()) + "#" + m.disc
	m.pending = &prAction{target: m.curTarget(), number: m.disc, label: label, discID: d.ID}
	switch key {
	case "x":
		if d.Closed {
			m.pending = nil
			return note("already closed — nothing to close")
		}
		cs := []choice{{label: "Resolved", value: "RESOLVED"}, {label: "Outdated", value: "OUTDATED"}, {label: "Duplicate", value: "DUPLICATE"}}
		return m.withDialog(newChoiceDialog("disc.close", "Close "+label+" as…", cs, false, nil)), nil
	case "X":
		if !d.Closed {
			m.pending = nil
			return note("already open — nothing to reopen")
		}
		p := m.pending
		m.pending = nil
		return m, func() tea.Msg {
			return prActionDoneMsg{text: "reopened " + p.label, err: reopenDisc(p.target, p.discID)}
		}
	case "L":
		if d.Locked {
			return m.withDialog(newConfirmDialog("disc.unlock", "Unlock "+label+"?", d.Title+" — people can comment on it again")), nil
		}
		return m.withDialog(newConfirmDialog("disc.lock", "Lock "+label+"?", d.Title+" — nobody can comment on it until it is unlocked")), nil
	}
	// A: the selected comment is, or stops being, the answer.
	e, _, has := doc.entry(st.sel)
	switch {
	case !has:
		m.pending = nil
		return note("select a comment first (n / N)")
	case e.parent != "":
		m.pending = nil
		return note("only a top-level comment can be the answer, not a reply")
	case !d.Answerable:
		m.pending = nil
		return note("the category " + orDash(d.Category) + " takes no answers")
	}
	m.pending.commentID = e.id
	for _, c := range d.Comments {
		if c.ID == e.id && c.IsAnswer {
			return m.withDialog(newConfirmDialog("disc.unmark", "Remove the answer mark in "+label+"?", e.author+"'s comment is no longer the answer")), nil
		}
	}
	return m.withDialog(newConfirmDialog("disc.mark", "Mark "+e.author+"'s comment as the answer of "+label+"?", d.Title)), nil
}
