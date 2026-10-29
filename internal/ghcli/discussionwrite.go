package ghcli

import (
	"errors"
	"strings"
)

const addDiscussionCommentMutation = `mutation($discussion:ID!,$body:String!,$reply:ID){addDiscussionComment(input:{discussionId:$discussion,body:$body,replyToId:$reply}){comment{id}}}`

// AddDiscussionComment adds a comment to a discussion; with replyToID it is a reply in that top-level
// comment's thread. The body travels as JSON on stdin, the ids as variables.
func AddDiscussionComment(host, discussionID, body, replyToID string) error {
	if strings.TrimSpace(discussionID) == "" {
		return errors.New("no discussion to comment on")
	}
	if strings.TrimSpace(body) == "" {
		return errors.New("an empty comment is not sent")
	}
	vars := map[string]any{"discussion": discussionID, "body": body}
	if replyToID != "" {
		vars["reply"] = replyToID
	}
	_, err := runWrite(host, addDiscussionCommentMutation, vars)
	return err
}

const (
	closeDiscussionMutation  = `mutation($discussion:ID!,$reason:DiscussionCloseReason){closeDiscussion(input:{discussionId:$discussion,reason:$reason}){discussion{id}}}`
	reopenDiscussionMutation = `mutation($discussion:ID!){reopenDiscussion(input:{discussionId:$discussion}){discussion{id}}}`
	markAnswerMutation       = `mutation($comment:ID!){markDiscussionCommentAsAnswer(input:{id:$comment}){discussion{id}}}`
	unmarkAnswerMutation     = `mutation($comment:ID!){unmarkDiscussionCommentAsAnswer(input:{id:$comment}){discussion{id}}}`
	lockMutation             = `mutation($id:ID!){lockLockable(input:{lockableId:$id}){lockedRecord{locked}}}`
	unlockMutation           = `mutation($id:ID!){unlockLockable(input:{lockableId:$id}){unlockedRecord{locked}}}`
)

// DiscCloseReasons are the reasons GitHub offers when a discussion is closed.
var DiscCloseReasons = []string{"RESOLVED", "OUTDATED", "DUPLICATE"}

func needID(id, what string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("no " + what)
	}
	return nil
}

// CloseDiscussion closes a discussion; reason is one of DiscCloseReasons, or "" for GitHub's default.
func CloseDiscussion(host, discussionID, reason string) error {
	if err := needID(discussionID, "discussion to close"); err != nil {
		return err
	}
	vars := map[string]any{"discussion": discussionID}
	if reason != "" {
		ok := false
		for _, r := range DiscCloseReasons {
			ok = ok || r == reason
		}
		if !ok {
			return errors.New("unknown close reason " + reason)
		}
		vars["reason"] = reason
	}
	_, err := runWrite(host, closeDiscussionMutation, vars)
	return err
}

// ReopenDiscussion reopens a closed discussion.
func ReopenDiscussion(host, discussionID string) error {
	if err := needID(discussionID, "discussion to reopen"); err != nil {
		return err
	}
	_, err := runWrite(host, reopenDiscussionMutation, map[string]any{"discussion": discussionID})
	return err
}

// MarkDiscussionAnswer marks a top-level comment as the answer of its discussion, or with mark false takes the mark away.
func MarkDiscussionAnswer(host, commentID string, mark bool) error {
	if err := needID(commentID, "comment to mark"); err != nil {
		return err
	}
	q := markAnswerMutation
	if !mark {
		q = unmarkAnswerMutation
	}
	_, err := runWrite(host, q, map[string]any{"comment": commentID})
	return err
}

// LockDiscussion locks a discussion (no more comments), or with lock false unlocks it.
func LockDiscussion(host, discussionID string, lock bool) error {
	if err := needID(discussionID, "discussion to lock"); err != nil {
		return err
	}
	q := lockMutation
	if !lock {
		q = unlockMutation
	}
	_, err := runWrite(host, q, map[string]any{"id": discussionID})
	return err
}
