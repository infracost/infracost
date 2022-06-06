include {
  path = find_in_parent_folders()
}

dependency "test" {
  config_path = "../prod"
}

dependency "test2" {
  config_path = "../prod2"
}

terraform {
  source = "..//modules/example"
}

inputs = {
  instance_type = dependency.test.outputs.aws_instance_type
  root_block_device_volume_size = 50
  block_device_volume_size = 100
  block_device_iops = dependency.test2.outputs.block_iops
  
  hello_world_function_memory_size = 512
}
