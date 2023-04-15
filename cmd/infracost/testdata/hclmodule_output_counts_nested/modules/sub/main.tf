variable "enabled" {
  type    = bool
  default = false
}

resource "aws_eip" "t" {
  count = var.enabled ? 1 : 0
}
