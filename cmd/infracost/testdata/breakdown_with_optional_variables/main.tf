terraform {
  experiments = [module_variable_optional_attrs]
}

variable "var1" {
  type = object({
    attr1 = string
    attr2 = optional(string)
  })

  default = {
    attr1 = "value"
  }
}

resource "aws_eip" "test" {
  count = coalesce(var.var1.attr2, "yes") == "yes" ? 1 : 0
}
