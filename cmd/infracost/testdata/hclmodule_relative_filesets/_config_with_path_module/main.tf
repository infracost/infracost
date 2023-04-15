variable "base_object" {
  default = {}
}

locals {
  loaded_files = [for v in ["config2.base.json"] : jsondecode(file("${path.module}/${v}"))]
}

output "result" {
  value = merge(concat([var.base_object], local.loaded_files)...)
}
