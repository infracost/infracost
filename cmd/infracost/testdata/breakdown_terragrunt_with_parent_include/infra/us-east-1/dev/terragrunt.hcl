include {
  path = find_in_parent_folders()
}

terraform {
  source = "..//modules/example"
}

inputs = {
  instance_type                 = "m5.4xlarge"
  root_block_device_volume_size = 50
  block_device_volume_size      = 100
  block_device_iops             = 800
}
