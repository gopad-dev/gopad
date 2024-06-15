[
  "alpha"
  "beta"
  "dev"
  "pre"
  "rc"
  "+incompatible"
  ] @keyword

(module_path) @string.special.url

(module_version) @string.special

(hash_version) @attribute

(hash) @string.special.symbol

[
  (number)
  (number_with_decimal)
  (hex_number)
  ] @constant.numeric

(checksum
  "go.mod" @string)

[
  ":"
  "."
  "-"
  "/"
  ] @punctuation.delimiter
