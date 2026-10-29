package ghcli

import (
	"encoding/json"
	"fmt"
	"strings"
)

// IssueFilter narrows the issue list. State is "open", "closed" or "all" ("" means open).
type IssueFilter struct {
	State                    string
	AssignedToMe, OpenedByMe bool
}

// IssueFilterArgs are the extra `gh issue list` flags for f. Only the three known states get
// through, so a stray value can never turn into another flag.
func IssueFilterArgs(f IssueFilter) []string {
	state := "open"
	if f.State == "closed" || f.State == "all" {
		state = f.State
	}
	a := []string{"--state", state}
	if f.AssignedToMe {
		a = append(a, "--assignee", "@me")
	}
	if f.OpenedByMe {
		a = append(a, "--author", "@me")
	}
	return a
}

// IssueDetail is what the issue screen shows.
type IssueDetail struct {
	Number                          int
	Title, Body, Author, State, URL string
	Labels, Assignees               []string
	Comments                        []Comment
	Locked, LockKnown               bool // whether the conversation is locked, and whether that could be read
}

// ParseIssueDetail reads the JSON of `gh issue view --json ...` (fields listed in FetchIssue).
func ParseIssueDetail(b []byte) (IssueDetail, error) {
	var raw struct {
		Number int       `json:"number"`
		Title  string    `json:"title"`
		Body   string    `json:"body"`
		Author *rawLogin `json:"author"`
		State  string    `json:"state"`
		URL    string    `json:"url"`
		Labels []struct {
			Name string `json:"name"`
		} `json:"labels"`
		Assignees []rawLogin `json:"assignees"`
		Comments  []struct {
			Author    *rawLogin `json:"author"`
			Body      string    `json:"body"`
			CreatedAt string    `json:"createdAt"`
		} `json:"comments"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return IssueDetail{}, err
	}
	d := IssueDetail{Number: raw.Number, Title: CleanText(raw.Title), Body: CleanText(raw.Body), Author: loginOrGhost(raw.Author),
		State: raw.State, URL: raw.URL}
	for _, l := range raw.Labels {
		d.Labels = append(d.Labels, CleanText(l.Name))
	}
	for _, a := range raw.Assignees {
		if a.Login != "" {
			d.Assignees = append(d.Assignees, CleanText(a.Login))
		}
	}
	for _, c := range raw.Comments {
		d.Comments = append(d.Comments, Comment{loginOrGhost(c.Author), CleanText(c.Body), c.CreatedAt})
	}
	return d, nil
}

// FetchIssue loads one issue with its comments.
func FetchIssue(r RepoRef, number string) (IssueDetail, error) {
	out, err := Run(Scoped(r, "issue", "view", number, "--json",
		"number,title,body,author,state,url,labels,assignees,comments")...)
	if err != nil {
		return IssueDetail{}, err
	}
	d, err := ParseIssueDetail(out)
	if err != nil {
		return d, err
	}
	if args := IssueLockedArgs(r, number); args != nil { // best effort: an old host or a missing scope leaves it unknown
		if b, lerr := Run(args...); lerr == nil {
			switch strings.TrimSpace(string(b)) {
			case "true":
				d.Locked, d.LockKnown = true, true
			case "false":
				d.LockKnown = true
			}
		}
	}
	return d, nil
}

// CommentIssueArgs is the `gh issue comment` command; the body is one --body=… argument.
func CommentIssueArgs(r RepoRef, number, body string) []string {
	return Scoped(r, "issue", "comment", number, "--body="+body)
}

// CommentIssue adds a comment to the issue.
func CommentIssue(r RepoRef, number, body string) error {
	_, err := Run(CommentIssueArgs(r, number, body)...)
	return err
}

// IssueCloseReasons are the reasons gh accepts when an issue is closed.
var IssueCloseReasons = []string{"completed", "not planned"}

// CloseIssueArgs is the `gh issue close` command. reason is one of IssueCloseReasons, or "" for gh's
// default; any other value yields nil so nothing unintended can be built from it.
func CloseIssueArgs(r RepoRef, number, reason string) []string {
	a := []string{"issue", "close", number}
	if reason != "" {
		ok := false
		for _, known := range IssueCloseReasons {
			ok = ok || known == reason
		}
		if !ok {
			return nil
		}
		a = append(a, "--reason="+reason)
	}
	return Scoped(r, a...)
}

// LockIssueArgs is `gh issue lock` (or `unlock`): a locked conversation takes no new comments.
func LockIssueArgs(r RepoRef, number string, lock bool) []string {
	sub := "lock"
	if !lock {
		sub = "unlock"
	}
	return Scoped(r, "issue", sub, number)
}

// IssueLockedArgs reads whether an issue is locked: `gh issue view` does not offer that field.
func IssueLockedArgs(r RepoRef, number string) []string {
	if number == "" || strings.Trim(number, "0123456789") != "" {
		return nil
	}
	endpoint, flags := r.APIEndpoint("/issues/" + number)
	return append(append([]string{"api"}, flags...), endpoint, "--jq", ".locked")
}

// ReopenIssueArgs is the `gh issue reopen` command.
func ReopenIssueArgs(r RepoRef, number string) []string {
	return Scoped(r, "issue", "reopen", number)
}

// EditIssueArgs is the `gh issue edit` command for label and assignee changes. Every name is
// its own --flag=value argument, so a name with spaces, an emoji, or a leading "-" stays one value.
// gh reads the label flags as comma-separated lists, so label names are quoted CSV-style when they
// hold a comma, a quote or a line break (user logins cannot).
func EditIssueArgs(r RepoRef, number string, addLabels, removeLabels, addAssignees, removeAssignees []string) []string {
	return editArgs("issue", r, number, addLabels, removeLabels, addAssignees, removeAssignees)
}

func editArgs(kind string, r RepoRef, number string, addLabels, removeLabels, addAssignees, removeAssignees []string) []string {
	a := []string{kind, "edit", number}
	for _, group := range []struct {
		flag  string
		names []string
		csv   bool
	}{
		{"--add-label=", addLabels, true}, {"--remove-label=", removeLabels, true},
		{"--add-assignee=", addAssignees, false}, {"--remove-assignee=", removeAssignees, false},
	} {
		for _, n := range group.names {
			if group.csv {
				n = csvQuote(n)
			}
			a = append(a, group.flag+n)
		}
	}
	return Scoped(r, a...)
}

// ParseLabelNames reads the JSON of `gh label list --json name`.
func ParseLabelNames(b []byte) ([]string, error) {
	var raw []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}
	var out []string
	for _, l := range raw {
		if l.Name != "" {
			out = append(out, l.Name)
		}
	}
	return out, nil
}

// ParseLoginLines reads one login per line, as printed by `gh api --jq '.[].login'`.
func ParseLoginLines(b []byte) []string {
	var out []string
	for _, l := range strings.Split(string(b), "\n") {
		if l = CleanText(strings.TrimSpace(l)); l != "" {
			out = append(out, l)
		}
	}
	return out
}

// ListLabels lists the repo's labels (up to 1000).
func ListLabels(r RepoRef) ([]string, error) {
	out, err := Run(Scoped(r, "label", "list", "--limit", "1000", "--json", "name")...)
	if err != nil {
		return nil, err
	}
	return ParseLabelNames(out)
}

// ListAssignees lists every user who can be assigned in the repo, page after page.
func ListAssignees(r RepoRef) ([]string, error) {
	endpoint, flags := r.APIEndpoint("/assignees?per_page=100")
	args := append(append([]string{"api", "--paginate"}, flags...), endpoint, "--jq", ".[].login")
	out, err := Run(args...)
	if err != nil {
		return nil, err
	}
	return ParseLoginLines(out), nil
}

// CloseIssue closes the issue, with a reason from IssueCloseReasons or "" for none.
func CloseIssue(r RepoRef, number, reason string) error {
	args := CloseIssueArgs(r, number, reason)
	if args == nil {
		return fmt.Errorf("unknown close reason %q", reason)
	}
	_, err := Run(args...)
	return err
}

// LockIssue locks the issue's conversation, or with lock false unlocks it.
func LockIssue(r RepoRef, number string, lock bool) error {
	_, err := Run(LockIssueArgs(r, number, lock)...)
	return err
}

// ReopenIssue reopens the issue.
func ReopenIssue(r RepoRef, number string) error {
	_, err := Run(ReopenIssueArgs(r, number)...)
	return err
}

// EditIssue changes the issue's labels and assignees.
func EditIssue(r RepoRef, number string, addLabels, removeLabels, addAssignees, removeAssignees []string) error {
	_, err := Run(EditIssueArgs(r, number, addLabels, removeLabels, addAssignees, removeAssignees)...)
	return err
}

// csvQuote quotes a value the way gh's comma-separated flags read it, when it needs it.
func csvQuote(name string) string {
	if !strings.ContainsAny(name, ",\"\n\r") {
		return name
	}
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
