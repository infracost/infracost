locals {
  enabled = var.enabled
}

variable "enabled" {
  type    = bool
  default = null
}

output "enabled" {
  value       = local.enabled
  description = "True if module is enabled, false otherwise"
}
