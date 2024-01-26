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

(module_path) @string.special.url

(file_path) @string

[
  (version)
  (go_version)
  ] @string.special
