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

(module_path) @string.special.path

(file_path) @string.special.path

[
  (version)
  (go_version)
  ] @string.special
