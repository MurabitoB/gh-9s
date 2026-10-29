package ghcli

import (
	"reflect"
	"testing"
)

func TestCloseAndReopenArgs(t *testing.T) {
	if got, want := CloseIssueArgs(acmeAPI, "14812", ""), []string{"issue", "close", "14812", "-R", "acme/api"}; !reflect.DeepEqual(got, want) {
		t.Errorf("close = %q; want %q", got, want)
	}
	if got, want := CloseIssueArgs(acmeAPI, "14812", "not planned"), []string{"issue", "close", "14812", "--reason=not planned", "-R", "acme/api"}; !reflect.DeepEqual(got, want) {
		t.Errorf("close as not planned = %q; want %q", got, want)
	}
	if got, want := CloseIssueArgs(acmeAPI, "14812", "completed"), []string{"issue", "close", "14812", "--reason=completed", "-R", "acme/api"}; !reflect.DeepEqual(got, want) {
		t.Errorf("close as completed = %q; want %q", got, want)
	}
	if got := CloseIssueArgs(acmeAPI, "14812", "--web"); got != nil {
		t.Errorf("an unknown reason must build no command, got %q", got)
	}
	if got, want := LockIssueArgs(acmeAPI, "7", true), []string{"issue", "lock", "7", "-R", "acme/api"}; !reflect.DeepEqual(got, want) {
		t.Errorf("lock = %q; want %q", got, want)
	}
	if got, want := LockIssueArgs(acmeAPI, "7", false), []string{"issue", "unlock", "7", "-R", "acme/api"}; !reflect.DeepEqual(got, want) {
		t.Errorf("unlock = %q; want %q", got, want)
	}
	if got, want := ReopenIssueArgs(RepoRef{}, "7"), []string{"issue", "reopen", "7"}; !reflect.DeepEqual(got, want) {
		t.Errorf("reopen = %q; want %q", got, want)
	}
}

func TestEditIssueArgs(t *testing.T) {
	got := EditIssueArgs(acmeAPI, "14812",
		[]string{"needs triage 🚧", "bug"}, []string{"wontfix"}, []string{"erin"}, []string{"frank"})
	want := []string{"issue", "edit", "14812",
		"--add-label=needs triage 🚧", "--add-label=bug", "--remove-label=wontfix",
		"--add-assignee=erin", "--remove-assignee=frank", "-R", "acme/api"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("EditIssueArgs =\n %q\nwant\n %q", got, want)
	}
}

func TestEditIssueArgsEmptyListsAddNothing(t *testing.T) {
	got := EditIssueArgs(acmeAPI, "1", nil, nil, nil, nil)
	if len(got) != 5 { // issue edit 1 -R acme/api
		t.Errorf("args = %q; empty change lists must add no flags", got)
	}
}

func TestEditIssueArgsKeepHostileNamesInOneArgument(t *testing.T) {
	for _, name := range []string{"--web", "-x", "a b; rm -rf ~", "$(id)"} {
		got := EditIssueArgs(acmeAPI, "1", []string{name}, nil, nil, nil)
		found := false
		for _, a := range got {
			if a == "--add-label="+name {
				found = true
			}
		}
		if !found {
			t.Errorf("%q was not kept as one --add-label= argument: %q", name, got)
		}
	}
}

// gh reads --add-label / --remove-label as comma-separated lists (CSV), so a name with a comma or
// a quote would be split or rejected unless it is quoted CSV-style.
func TestEditIssueArgsQuoteLabelsThatGhWouldSplit(t *testing.T) {
	got := EditIssueArgs(acmeAPI, "1", []string{"a,b", `say "hi"`, "multi\nline", "plain"}, []string{"x,y"}, nil, nil)
	want := []string{"issue", "edit", "1",
		`--add-label="a,b"`, `--add-label="say ""hi"""`, "--add-label=\"multi\nline\"", "--add-label=plain",
		`--remove-label="x,y"`, "-R", "acme/api"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got\n %q\nwant\n %q", got, want)
	}
}

func TestParseLabelNamesAndLogins(t *testing.T) {
	labels, err := ParseLabelNames([]byte(`[{"name":"bug"},{"name":"needs triage 🚧"},{"name":""}]`))
	if err != nil || !reflect.DeepEqual(labels, []string{"bug", "needs triage 🚧"}) {
		t.Errorf("labels = %q err %v; want the two names, the empty one dropped", labels, err)
	}
	logins := ParseLoginLines([]byte("erin\nfrank\n\n  gina  \n"))
	if !reflect.DeepEqual(logins, []string{"erin", "frank", "gina"}) {
		t.Errorf("logins = %q; want blank lines dropped and names trimmed", logins)
	}
	if got, err := ParseLabelNames([]byte(`[]`)); err != nil || len(got) != 0 {
		t.Errorf("empty list: %v %v", got, err)
	}
}

func TestIssueLockedArgsReadTheRestFieldAndRefuseABadNumber(t *testing.T) {
	got := IssueLockedArgs(RepoRef{Host: "ghe.example.com", Owner: "acme", Name: "api"}, "14812")
	want := []string{"api", "--hostname", "ghe.example.com", "repos/acme/api/issues/14812", "--jq", ".locked"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("IssueLockedArgs = %q; want %q", got, want)
	}
	if IssueLockedArgs(acmeAPI, "../x") != nil || IssueLockedArgs(acmeAPI, "") != nil {
		t.Error("a number that is not digits must build no command")
	}
}
