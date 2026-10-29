package tui

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/MurabitoB/gh-9s/internal/ghcli"
)

type prStateFakes struct {
	closes, reopens []string
	readies, drafts []string
	edits           []editCall
	labels          []string
	assignees       []string
	err             error
}

func installPRState(t *testing.T) *prStateFakes {
	t.Helper()
	f := &prStateFakes{labels: []string{"bug", "docs", "wontfix"}, assignees: []string{"erin", "frank"}}
	oldClose, oldReopen, oldReady, oldEdit := closePR, reopenPR, readyPR, editPR
	oldLabels, oldAssignees := listLabels, listAssignees
	closePR = func(r ghcli.RepoRef, n string) error { f.closes = append(f.closes, n); return f.err }
	reopenPR = func(r ghcli.RepoRef, n string) error { f.reopens = append(f.reopens, n); return f.err }
	readyPR = func(r ghcli.RepoRef, n string, undo bool) error {
		if undo {
			f.drafts = append(f.drafts, n)
		} else {
			f.readies = append(f.readies, n)
		}
		return f.err
	}
	editPR = func(r ghcli.RepoRef, n string, al, rl, aa, ra []string) error {
		f.edits = append(f.edits, editCall{r, n, al, rl, aa, ra})
		return f.err
	}
	listLabels = func(ghcli.RepoRef) ([]string, error) { return f.labels, nil }
	listAssignees = func(ghcli.RepoRef) ([]string, error) { return f.assignees, nil }
	t.Cleanup(func() {
		closePR, reopenPR, readyPR, editPR = oldClose, oldReopen, oldReady, oldEdit
		listLabels, listAssignees = oldLabels, oldAssignees
	})
	return f
}

func prDetailScreen(t *testing.T, d ghcli.PRDetail) model {
	t.Helper()
	installPRFetch(t, d)
	m := prListModel()
	m.target, m.prTarget = acme, acme
	return openDetail(t, m)
}

func withState(state string, draft bool) ghcli.PRDetail {
	d := fixturePR()
	d.State, d.IsDraft = state, draft
	d.Assignees = []string{"erin"}
	return d
}

func TestXAsksBeforeClosingAPRAndNamesIt(t *testing.T) {
	f := installPRState(t)
	m := keyAll(prDetailScreen(t, withState("OPEN", false)), "x")
	if m.dlg == nil || m.dlg.kind != dialogConfirm || !strings.Contains(m.View(), "acme/api#482") || !strings.Contains(m.View(), "Fix flaky retry in deploy") {
		t.Fatalf("dialog = %+v; want a confirmation naming acme/api#482 and its title", m.dlg)
	}
	if len(f.closes) != 0 {
		t.Fatal("nothing may be closed before the confirmation")
	}
	if declined := keyAll(m, "n"); len(f.closes) != 0 || declined.dlg != nil {
		t.Errorf("declining must close nothing: %v", f.closes)
	}
	m = keyAll(m, "y")
	if !reflect.DeepEqual(f.closes, []string{"482"}) || !strings.Contains(m.status, "closed") {
		t.Errorf("closes=%v status=%q", f.closes, m.status)
	}
}

func TestXOnAClosedOrMergedPRSaysSo(t *testing.T) {
	for state, want := range map[string]string{"CLOSED": "already closed", "MERGED": "merged"} {
		f := installPRState(t)
		m := keyAll(prDetailScreen(t, withState(state, false)), "x")
		if m.dlg != nil || len(f.closes) != 0 || !strings.Contains(m.status, want) {
			t.Errorf("%s: dialog=%v closes=%v status=%q; want a note containing %q", state, m.dlg != nil, f.closes, m.status, want)
		}
	}
}

func TestReopenNeedsNoConfirmationAndWorksOnlyOnAClosedPR(t *testing.T) {
	f := installPRState(t)
	m := keyAll(prDetailScreen(t, withState("CLOSED", false)), "X")
	if m.dlg != nil || !reflect.DeepEqual(f.reopens, []string{"482"}) || !strings.Contains(m.status, "reopened") {
		t.Errorf("dialog=%v reopens=%v status=%q", m.dlg != nil, f.reopens, m.status)
	}
	for state, want := range map[string]string{"OPEN": "already open", "MERGED": "merged"} {
		f := installPRState(t)
		m := keyAll(prDetailScreen(t, withState(state, false)), "X")
		if len(f.reopens) != 0 || !strings.Contains(m.status, want) {
			t.Errorf("%s: reopens=%v status=%q; want a note containing %q", state, f.reopens, m.status, want)
		}
	}
}

func TestAFailedCloseIsShownAndNothingElseHappens(t *testing.T) {
	f := installPRState(t)
	f.err = errors.New("HTTP 403: forbidden")
	m := keyAll(prDetailScreen(t, withState("OPEN", false)), "x", "y")
	if !strings.Contains(m.status, "✗") || !strings.Contains(m.status, "403") {
		t.Errorf("status = %q; want the error", m.status)
	}
}

func TestDMarksADraftReadyAfterAskingAndTurnsAReadyPRIntoADraft(t *testing.T) {
	f := installPRState(t)
	m := keyAll(prDetailScreen(t, withState("OPEN", true)), "D")
	if m.dlg == nil || !strings.Contains(m.View(), "ready for review") || len(f.readies) != 0 {
		t.Fatalf("want a question about marking it ready, nothing sent yet; dialog=%v readies=%v", m.dlg != nil, f.readies)
	}
	m = keyAll(m, "y")
	if !reflect.DeepEqual(f.readies, []string{"482"}) || len(f.drafts) != 0 || !strings.Contains(m.status, "ready for review") {
		t.Errorf("readies=%v drafts=%v status=%q", f.readies, f.drafts, m.status)
	}

	f = installPRState(t)
	m = keyAll(prDetailScreen(t, withState("OPEN", false)), "D")
	if m.dlg == nil || !strings.Contains(m.View(), "draft") {
		t.Fatalf("want a question about turning it into a draft; dialog=%v", m.dlg != nil)
	}
	m = keyAll(m, "y")
	if !reflect.DeepEqual(f.drafts, []string{"482"}) || len(f.readies) != 0 {
		t.Errorf("readies=%v drafts=%v", f.readies, f.drafts)
	}
}

func TestDOnAClosedPRSaysItNeedsAnOpenPR(t *testing.T) {
	f := installPRState(t)
	m := keyAll(prDetailScreen(t, withState("CLOSED", false)), "D")
	if m.dlg != nil || len(f.readies)+len(f.drafts) != 0 || !strings.Contains(m.status, "open") {
		t.Errorf("dialog=%v status=%q", m.dlg != nil, m.status)
	}
}

func TestPRLabelsKeyListsRepoLabelsWithCurrentOnesChecked(t *testing.T) {
	f := installPRState(t)
	m := keyAll(prDetailScreen(t, withState("OPEN", false)), "l")
	v := m.View()
	for _, want := range []string{"[x] bug", "[ ] docs", "[ ] wontfix"} { // fixturePR is labelled "bug"
		if !strings.Contains(v, want) {
			t.Errorf("label dialog lacks %q:\n%s", want, v)
		}
	}
	m = keyAll(m, "space", "down", "down", "space", "enter") // untick bug, tick wontfix
	want := editCall{repo: acme, number: "482", addLabels: []string{"wontfix"}, removeLabels: []string{"bug"}}
	if len(f.edits) != 1 || !reflect.DeepEqual(f.edits[0], want) {
		t.Errorf("edits = %+v; want %+v", f.edits, want)
	}
}

func TestPRAssigneesKeyChangesAssignees(t *testing.T) {
	f := installPRState(t)
	m := keyAll(prDetailScreen(t, withState("OPEN", false)), "s")
	if v := m.View(); !strings.Contains(v, "[x] erin") || !strings.Contains(v, "[ ] frank") {
		t.Fatalf("assignee dialog:\n%s", v)
	}
	keyAll(m, "down", "space", "enter")
	want := editCall{repo: acme, number: "482", addAssignees: []string{"frank"}}
	if len(f.edits) != 1 || !reflect.DeepEqual(f.edits[0], want) {
		t.Errorf("edits = %+v; want %+v", f.edits, want)
	}
}

func TestPRDetailFooterAndHelpNameTheNewKeys(t *testing.T) {
	m := prDetailScreen(t, withState("OPEN", false))
	if v := m.View(); !strings.Contains(v, "x: close") || !strings.Contains(v, "D: draft") {
		t.Errorf("footer lacks the new keys:\n%s", v)
	}
}
