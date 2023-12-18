include {
  path = find_in_parent_folders()
}

terraform {
  source = "..//modules/example"
}

dependency "dev" {
  # This path does not exist and will cause Terragrunt to exit creating a "stack".
  # This means that we return early from the Terragrunt hcl parser and won't properly
  # determine the Terragrunt project structure. This leans to mismatches in prior/current
  # projects when using the diff command.
  config_path = "../../foo"
}

inputs = {
  instance_type = "t2.micro"
  root_block_device_volume_size = 50
  block_device_volume_size = 100
  block_device_iops = 400
  
  hello_world_function_memory_size = 512
}
