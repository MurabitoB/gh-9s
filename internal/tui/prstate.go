package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/MurabitoB/gh-9s/internal/ghcli"
)

// prStateKey handles the keys of the PR detail that change the PR itself: x close, X reopen,
// D draft / ready for review, U update the branch, l labels, s assignees. Each says why when it does not apply.
func (m model) prStateKey(key string, d ghcli.PRDetail) (model, tea.Cmd) {
	note := func(text string) (model, tea.Cmd) {
		m.status = styDim.Render(text)
		return m, nil
	}
	if d.State == "" && key != "l" && key != "s" {
		return note("still loading the PR")
	}
	switch key {
	case "x":
		switch d.State {
		case "CLOSED":
			return note("already closed — nothing to close")
		case "MERGED":
			return note("already merged — nothing to close")
		}
		m, p := m.startPRAction()
		return m.withDialog(newConfirmDialog("pr.close", "Close "+p.label+"?", d.Title+" — closed without being merged")), nil
	case "X":
		switch d.State {
		case "OPEN":
			return note("already open — nothing to reopen")
		case "MERGED":
			return note("a merged PR cannot be reopened")
		}
		m, p := m.startPRAction()
		m.pending = nil
		return m, func() tea.Msg {
			return prActionDoneMsg{text: "reopened " + p.label, err: reopenPR(p.target, p.number)}
		}
	case "D":
		if d.State != "OPEN" {
			return note("only an open PR can be a draft or ready for review")
		}
		m, p := m.startPRAction()
		if d.IsDraft {
			return m.withDialog(newConfirmDialog("pr.ready", "Mark "+p.label+" ready for review?", d.Title+" — reviewers are notified")), nil
		}
		return m.withDialog(newConfirmDialog("pr.draft", "Turn "+p.label+" into a draft?", d.Title)), nil
	case "U":
		if d.State != "OPEN" {
			return note("only an open PR's branch can be updated")
		}
		m, p := m.startPRAction()
		return m.withDialog(newConfirmDialog("pr.update", "Update "+p.label+"?", "Merge "+orDash(d.BaseRef)+" into "+orDash(d.HeadRef)+": adds a commit to the PR's branch")), nil
	}
	// l, s: fetch the repo's labels or assignable users, then show them with the PR's own ticked.
	m, p := m.startPRAction()
	kind := "labels"
	p.current = d.Labels
	if key == "s" {
		kind, p.current = "assignees", d.Assignees
	}
	m.status = styDim.Render("loading " + kind + "…")
	tg, id := p.target, m.itemID(m.pr)
	p.want = kind
	return m, func() tea.Msg {
		var names []string
		var err error
		if kind == "labels" {
			names, err = listLabels(tg)
		} else {
			names, err = listAssignees(tg)
		}
		return choicesMsg{view: viewPRDetail, id: id, kind: kind, names: names, err: err}
	}
}

func init() {
	done := func(text string, do func(p *prAction) error) func(model, dialogResultMsg) (model, tea.Cmd) {
		return func(m model, r dialogResultMsg) (model, tea.Cmd) {
			p := m.pending
			m.pending = nil
			if !r.ok || p == nil {
				return m, nil
			}
			return m, func() tea.Msg { return prActionDoneMsg{text: text + " " + p.label, err: do(p)} }
		}
	}
	dialogHandlers["pr.close"] = done("closed", func(p *prAction) error { return closePR(p.target, p.number) })
	dialogHandlers["pr.ready"] = done("marked ready for review:", func(p *prAction) error { return readyPR(p.target, p.number, false) })
	dialogHandlers["pr.update"] = done("updated the branch of", func(p *prAction) error { return updatePRBranch(p.target, p.number) })
	dialogHandlers["pr.draft"] = done("turned into a draft:", func(p *prAction) error { return readyPR(p.target, p.number, true) })
	edit := func(what string, labels bool) func(model, dialogResultMsg) (model, tea.Cmd) {
		return func(m model, r dialogResultMsg) (model, tea.Cmd) {
			p := m.pending
			m.pending = nil
			if !r.ok || p == nil {
				return m, nil
			}
			add, remove := diffNames(p.current, r.values)
			if len(add)+len(remove) == 0 {
				m.status = styDim.Render("no " + what + " changes")
				return m, nil
			}
			return m, func() tea.Msg {
				var err error
				if labels {
					err = editPR(p.target, p.number, add, remove, nil, nil)
				} else {
					err = editPR(p.target, p.number, nil, nil, add, remove)
				}
				return prActionDoneMsg{text: "updated " + what + " on " + p.label, err: err}
			}
		}
	}
	dialogHandlers["pr.labels"] = edit("labels", true)
	dialogHandlers["pr.assignees"] = edit("assignees", false)
}
