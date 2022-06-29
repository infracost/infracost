include {
  path = find_in_parent_folders()
}

locals {
  local_config = {
    root_block_device_volume_size = 100
    block_device_volume_size = 1000
    block_device_iops = 800

    hello_world_function_memory_size = 1024
  }

  loaded_config = yamldecode(file("./config.yml"))
  config = merge(local.local_config, local.loaded_config)
}

terraform {
  source = "..//modules/example2"
}

inputs = local.config
