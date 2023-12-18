terraform {
  required_version = ">= 0.13.0"
}

locals {
  content = var.condition ? "" : SEE_ABOVE_ERROR_MESSAGE(true ? null : "ERROR: ${var.error_message}")
}
