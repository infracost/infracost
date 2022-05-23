include {
  path = find_in_parent_folders()
}

dependency "ter" {
  config_path = "..//prod"
}

terraform {
  source = "..//modules/example"
}

inputs = {
  instance_type = dependency.ter.outputs.aws_instance_type
  root_block_device_volume_size = 50
  block_device_volume_size = 100
  block_device_iops = 400
  
  hello_world_function_memory_size = 512
}
