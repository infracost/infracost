include {
  path = find_in_parent_folders()
}

terraform {
  source = "..//modules/example"
}

locals {
  test = yamldecode(sops_decrypt_file("test.yml"))
  test2 = jsondecode(sops_decrypt_file("test.json"))
  test3 = run_cmd("echo", "hello")
}

inputs = {
  instance_type = "m5.4xlarge"
  root_block_device_volume_size = 100
  block_device_volume_size = 1000
  block_device_iops = 800
  
  hello_world_function_memory_size = 1024
}
