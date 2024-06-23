((comment) @injection.content
 (#set! injection.language "comment"))

(call_expression
  (selector_expression) @_function
  (#any-of? @_function "regexp.Match" "regexp.MatchReader" "regexp.MatchString" "regexp.Compile" "regexp.CompilePOSIX" "regexp.MustCompile" "regexp.MustCompilePOSIX")
  (argument_list
    .
    [
      (raw_string_literal)
      (interpreted_string_literal)
    ] @injection.content
    (#set! injection.language "regex")))

((call_expression
   function: (selector_expression
               field: (field_identifier) @_method)
   arguments: (argument_list
                .
                (interpreted_string_literal) @injection.content))
  (#any-of? @_method "Printf" "Sprintf" "Fatalf" "Scanf" "Errorf" "Skipf" "Logf")
  (#set! injection.language "printf"))

((call_expression
   function: (selector_expression
               field: (field_identifier) @_method)
   arguments: (argument_list
                (_)
                .
                (interpreted_string_literal) @injection.content))
  (#any-of? @_method "Fprintf" "Fscanf" "Appendf" "Sscanf")
  (#set! injection.language "printf"))

(call_expression
  (selector_expression) @_function
  (#match? @_function "^[a-zA-Z_][a-zA-Z0-9_]*.(Query|Exec|Prepare|Get|Select)(Context)?$")
  (argument_list
    .
    [
      (raw_string_literal)
      (interpreted_string_literal)
    ] @injection.content
	(#offset! @injection.content 0 1 0 -1)
    (#set! injection.language "sql")))
