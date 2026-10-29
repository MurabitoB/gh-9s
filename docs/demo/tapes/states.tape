Output states.gif
Source tapes/common.tape

Type "gh-9s"
Enter
Sleep 2s

# pull requests: closed, then merged, then back to the open ones
Type "1"
Sleep 2s
Type "C"
Sleep 2500ms
Type "m"
Sleep 2500ms
Type "O"
Sleep 2s

# a pull request: closing asks first, nothing is sent until y
Enter
Sleep 3s
Type "x"
Sleep 3s
Escape
Sleep 1s

# a draft goes to review only after the question
Type "D"
Sleep 3s
Escape
Sleep 1s
Escape
Sleep 1s

# discussions: the closed ones, then only the unanswered
Type "5"
Sleep 2500ms
Type "C"
Sleep 2500ms
Type "O"
Sleep 1500ms
Type "u"
Sleep 2500ms
Type "u"
Sleep 1s

# one discussion: close with a reason, mark the answer, lock
Enter
Sleep 3s
Type "x"
Sleep 3s
Escape
Sleep 1s
Type "n"
Sleep 500ms
Type "n"
Sleep 500ms
Type "n"
Sleep 500ms
Type "n"
Sleep 800ms
Type "A"
Sleep 3s
Escape
Sleep 1s
Type "L"
Sleep 3s
Escape
Sleep 1s
Type "q"
Sleep 500ms
