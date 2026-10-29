---
title: Issues
description: List, filter, read, comment on, close, lock and label issues.
---

The issue view (`2` or `:issues`) lists issues in the current repository. Press `enter` to open one.

## List

| Key | Action |
|---|---|
| `enter` | Open the issue |
| `a` | Send the issue to a coding agent (asks first) |
| `O` / `C` / `A` | Show open / closed / all issues |
| `u` | Show only issues assigned to you |
| `m` | Show only issues you opened |
| `f` | Pin or unpin this repository |
| `o` | Open in the browser |

On a terminal 120 columns or wider, the list has a preview pane on the right, the same as the PR list. See [Pull requests](/gh-9s/guides/pull-requests/#preview-pane).

## Detail screen

Press `enter` on an issue to read its description and discussion. On a terminal about 100 columns of text wide or more, a sidebar on the right shows the labels and assignees.

| Key | Action |
|---|---|
| `a` | Send the issue to a coding agent (asks first) |
| `C` | Write a comment |
| `x` | Close the issue, as completed or not planned |
| `X` | Reopen the issue |
| `L` | Lock or unlock the conversation (asks first) |
| `l` | Edit labels |
| `s` | Edit assignees |
| `tab` | Select an image |
| `enter` | Show the selected image full screen |
| `o` | Open in the browser |

A locked conversation shows `locked` under the title. If gh-9s cannot tell whether it is locked (an older host, or a token without access), `L` says so and changes nothing.

`a` starts a coding agent on the issue in Orca, herdr or tmux. See [Send an issue to an agent](/gh-9s/guides/send-to-agent/).

Long comments can be written in your `$EDITOR` with `ctrl+e`. See [Images and Markdown](/gh-9s/guides/images-and-markdown/) for how text and images are rendered.
