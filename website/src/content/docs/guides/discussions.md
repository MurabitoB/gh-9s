---
title: Discussions
description: List a repository's Discussions by category and state, read them with their replies, comment, reply, close, mark an answer and lock.
---

The Discussions view (`5` or `:discussions`, short form `:disc`) lists the Discussions of the current repository, the ones with the latest activity first. Press `enter` to open one. If the repository has Discussions turned off, the view says so.

## List

| Key | Action |
|---|---|
| `enter` | Open the discussion |
| `O` / `C` / `A` | Show open / closed / all discussions |
| `u` / `a` | Show only unanswered / only answered ones (press again for both) |
| `m` | Read more discussions |
| `c` | Narrow the list to one category, or back to all |
| `o` | Open in the browser |

A `✓` before the title marks a discussion that has an answer; `[closed]` and `[locked]` are marked too. The list reads open discussions, 50 at a time. When there are more, the status line says "first 50 of N" and `m` reads 50 more, up to 500; the filters and the category narrow the list on GitHub's side, so the rows you read are always of what you asked for. A new filter starts again at the first page. Switching repository forgets the category and the filters. The title bar names an active filter.

On a terminal 120 columns or wider the selected discussion's text is shown in a pane on the right, like the issue list.

## One discussion

The screen shows the text, then the comments, each with its replies under it (GitHub nests replies one level deep). The comment marked as the answer shows `✓ answer`. Up to 100 comments and 50 replies per comment are read; when there are more, a line says how many are not shown, and `o` opens the discussion in the browser.

| Key | Action |
|---|---|
| `j` / `k` | Scroll |
| `n` / `N` | Select the next / previous comment or reply (marked `▌` in the margin) |
| `C` | Write a new comment (asks first) |
| `r` | Reply to the selected comment (asks first); a reply to a reply goes into the same thread |
| `x` | Close, as resolved, outdated or duplicate |
| `X` | Reopen |
| `A` | Mark the selected comment as the answer, or take the mark away (asks first) |
| `L` | Lock or unlock the discussion (asks first) |
| `o` | Open in the browser |
| `R` | Read it again |

The question for a comment or reply names the repository and the number, and a reply names whom it answers. A locked discussion cannot be written to, and the screen says so instead of asking. Everything GitHub sends is cleaned of control characters before it is drawn.

Reading and writing use the `repo` permission that `gh` has by default; no extra scope is needed. Enterprise hosts work the same way as everywhere else in gh-9s.

`A` works on a top-level comment in a category that takes answers (such as Q&A), not on a reply. Closing, marking an answer and locking need write access to the discussion, as on GitHub; when GitHub refuses, the status line shows its message.

Not in this release: starting or editing a discussion, voting, and selecting images in the text.
