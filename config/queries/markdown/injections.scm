(fenced_code_block
  (info_string
    (language) @_lang)
  (code_fence_content) @injection.content
  (#set-lang-from-info-string! @_lang))

([
   (inline)
   (pipe_table_cell)
   ] @injection.content
  (#set! injection.language "markdown-inline"))
