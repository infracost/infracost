include "root" {
  path = find_in_parent_folders()
}

include "common" {
  path   = "${dirname(find_in_parent_folders())}/common/common.hcl"
  expose = true
}

terraform {
  source = "..//modules/example"
}

inputs = {
  instance_type                 = "m5.4xlarge"
  root_block_device_volume_size = 100
  block_device_volume_size      = 1000
  block_device_iops             = 800
}
