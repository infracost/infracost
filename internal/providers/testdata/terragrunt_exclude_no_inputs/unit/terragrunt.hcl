# The exclude block (Terragrunt >= 0.73) is unknown to the vendored Terragrunt parser, so decoding
# this file produces a diagnostic. There is no inputs attribute for DiagnosticsFunc to recover.
exclude {
  if      = true
  actions = ["all"]
}
