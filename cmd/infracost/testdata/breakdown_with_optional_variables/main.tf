terraform {
  experiments = [module_variable_optional_attrs]
}

variable "var1" {
  type = object({
    attr1 = string
    attr2 = optional(string)
  })

  default = {
    attr1 = "value1"
  }
}

variable "var2" {
  type = object({
    attr1 = string
    attr2 = optional(string)
  })

  default = {
    attr1 = "blank"
  }
}

resource "aws_eip" "test1" {
  count = var.var1.attr1 == "value1" ? 2 : 1
}

resource "aws_eip" "test2" {
  count = var.var2.attr1 == "value2" ? 2 : 1
}
