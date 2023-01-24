include {
  path = find_in_parent_folders()
}

terraform {
  source = "..//modules/example"
}

dependency "core" {
  config_path = "../core"
}

inputs = {
  resource_group_name = dependency.core.outputs.resource_group_name
}
