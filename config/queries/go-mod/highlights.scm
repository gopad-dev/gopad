[
  "require"
  "replace"
  "go"
  "toolchain"
  "exclude"
  "retract"
  "module"
  ] @keyword

"=>" @operator

[
  "("
  ")"
  ] @punctuation.bracket

(comment) @comment @spell

(module_path) @string

(file_path) @string

[
  (version)
  (go_version)
  ] @string.special
