terraform {
  experiments = [module_variable_optional_attrs]
}

variable "var1" {
  type = object({
    attr1 = string
    attr2 = optional(string)
    attr3 = optional(string, "value3")
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

variable "var3" {
  type = list(object({
    attr1 = string
    attr2 = optional(string)
    attr3 = optional(string, "value3")
  }))

  default = [{
    attr1 = "value1"
  }]
}

resource "aws_eip" "test1" {
  count = var.var1.attr1 == "value1" ? 2 : 1
}

resource "aws_eip" "test2" {
  count = var.var2.attr1 == "value2" ? 2 : 1
}

resource "aws_eip" "test3" {
  count = var.var1.attr3 == "value3" ? 2 : 1
}

resource "aws_eip" "test4" {
  count = var.var3[0].attr3 == "value3" ? 2 : 1
}

