terraform {
  source = "git::https://github.com/infracost/infracost.git//examples/terragrunt/modules/example?ref=74725d6e91aa91d7283642b7ed3316d12f271212"
}

inputs = {
  instance_type = "m5.4xlarge"
  root_block_device_volume_size = 100
  block_device_volume_size = 1000
  block_device_iops = 800

  hello_world_function_memory_size = 1024
}
