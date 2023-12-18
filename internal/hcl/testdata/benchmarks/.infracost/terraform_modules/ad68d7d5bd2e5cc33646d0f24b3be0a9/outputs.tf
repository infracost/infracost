output "checked" {
  description = "Whether the condition has been checked (used for assertion dependencies)."
  value       = local.content == "" ? true : true
}
