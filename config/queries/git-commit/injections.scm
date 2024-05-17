(((scissors)
   (message) @injection.content)
  (#set! injection.language "diff"))

((rebase_command) @injection.content
  (#set! injection.language "git-rebase"))
