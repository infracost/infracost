terraform {
  source = "git::https://github.com/infracost/infracost.git//examples/terragrunt/modules/example?ref=1f94a2fd07b3e29deea4706b5d2fdc68c1d02aad"
}

inputs = {
  instance_type = "m5.8xlarge"
  root_block_device_volume_size = 100
  block_device_volume_size = 1000
  block_device_iops = 800

  hello_world_function_memory_size = 1024
}
