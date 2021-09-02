include {
  path = find_in_parent_folders()
}

terraform {
  source = "..//modules/example"
}

inputs = {
  instance_type = "t2.micro"
  root_block_device_volume_size = 50
  block_device_volume_size = 100
  block_device_iops = 400
  
  hello_world_function_memory_size = 512
}
