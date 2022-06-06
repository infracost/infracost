include {
  path = find_in_parent_folders()
}

terraform {
  source = "..//modules/example2"
}

inputs = {
  instance_type = "m5.4xlarge"
  root_block_device_volume_size = 100
  block_device_volume_size = 1000
  block_device_iops = 800
  
  hello_world_function_memory_size = 1024
}
