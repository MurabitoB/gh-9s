package ghcli

// ClosePRArgs is the `gh pr close` command.
func ClosePRArgs(r RepoRef, number string) []string { return Scoped(r, "pr", "close", number) }

// ReopenPRArgs is the `gh pr reopen` command.
func ReopenPRArgs(r RepoRef, number string) []string { return Scoped(r, "pr", "reopen", number) }

// ReadyPRArgs is `gh pr ready`, which marks a draft ready for review; with undo it turns a ready PR back into a draft.
func ReadyPRArgs(r RepoRef, number string, undo bool) []string {
	a := []string{"pr", "ready", number}
	if undo {
		a = append(a, "--undo")
	}
	return Scoped(r, a...)
}

// EditPRArgs is the `gh pr edit` command for label and assignee changes (see EditIssueArgs).
func EditPRArgs(r RepoRef, number string, addLabels, removeLabels, addAssignees, removeAssignees []string) []string {
	return editArgs("pr", r, number, addLabels, removeLabels, addAssignees, removeAssignees)
}

// ClosePR closes the PR without merging it.
func ClosePR(r RepoRef, number string) error {
	_, err := Run(ClosePRArgs(r, number)...)
	return err
}

// ReopenPR reopens a closed PR.
func ReopenPR(r RepoRef, number string) error {
	_, err := Run(ReopenPRArgs(r, number)...)
	return err
}

// ReadyPR marks a draft PR ready for review, or with undo turns a PR back into a draft.
func ReadyPR(r RepoRef, number string, undo bool) error {
	_, err := Run(ReadyPRArgs(r, number, undo)...)
	return err
}

// EditPR changes the PR's labels and assignees.
func EditPR(r RepoRef, number string, addLabels, removeLabels, addAssignees, removeAssignees []string) error {
	_, err := Run(EditPRArgs(r, number, addLabels, removeLabels, addAssignees, removeAssignees)...)
	return err
}
