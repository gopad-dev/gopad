((comment) @injection.content
 (#set! injection.language "comment"))

((element_comment) @injection.content
  (#set! injection.language "comment"))

((script_block_text) @injection.content
  (#set! injection.language "javascript"))

((script_element_text) @injection.content
  (#set! injection.language "javascript"))

((style_element_text) @injection.content
  (#set! injection.language "css"))