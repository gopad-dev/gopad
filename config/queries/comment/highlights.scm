; TODO level tags
((tag (name) @comment.todo)
  (#any-of? @comment.todo "TODO"))

("text" @comment.todo
  (#any-of? @comment.todo "TODO" "HINT" "MARK" "PASSED" "STUB" "MOCK"))

; Note level tags
((tag (name) @comment.note)
  (#any-of? @comment.note "NOTE" "INFO" "PERF" "OPTIMIZE" "PERFORMANCE" "QUESTION" "ASK"))

("text" @comment.note
    (#any-of? @comment.note "NOTE" "INFO" "PERF" "OPTIMIZE" "PERFORMANCE" "QUESTION" "ASK"))

; Warning level tags
((tag (name) @comment.warning)
  (#any-of? @comment.warning "WARNING" "WARN" "HACK" "TEST" "TEMP"))

("text" @comment.warning
  (#any-of? @comment.warning "WARNING" "WARN" "HACK" "TEST" "TEMP"))

; Error level tags
((tag (name) @comment.error)
  (#any-of? @comment.error "ERROR" "BUG" "FIXME" "ISSUE" "XXX" "FIX" "SAFETY" "FIXIT" "FAILED" "DEBUG"))

("text" @comment.error
  (#any-of? @comment.error "ERROR" "BUG" "FIXME" "ISSUE" "XXX" "FIX" "SAFETY" "FIXIT" "FAILED" "DEBUG"))

(uri) @markup.link.url
