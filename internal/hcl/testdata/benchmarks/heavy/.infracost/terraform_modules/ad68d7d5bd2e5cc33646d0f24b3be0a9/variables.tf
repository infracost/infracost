variable "condition" {
  description = "The condition to ensure is `true`."
  type        = bool
}

variable "error_message" {
  description = "The error message to display if the condition evaluates to `false`."
  type        = string
}
