variable "base_object" {
  default = {}
}

locals {
  loaded_files = [for v in ["config3.base.json"] : jsondecode(file("${path.cwd}/_config_with_path_cwd/${v}"))]
}

output "result" {
  value = merge(concat([var.base_object], local.loaded_files)...)
}
