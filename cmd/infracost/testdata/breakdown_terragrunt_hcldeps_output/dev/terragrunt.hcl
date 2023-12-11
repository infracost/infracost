include {
  path = find_in_parent_folders()
}

dependency "test" {
  config_path = "../prod"
}

dependency "test2" {
  config_path = "../prod2"
}

dependency "stag-dep" {
  config_path = "../stag"
}

terraform {
  source = "..//modules/example"
}

locals {
  region = "wa"
}

inputs = {
  instance_type                 = dependency.test.outputs.aws_instance_type
  root_block_device_volume_size = 50
  block_device_volume_size      = 100
  block_device_iops             = dependency.test2.outputs.block_iops

  bad  = dependency.test2.outputs.map_existing["test"][local.region].id
  bad2 = get_env("NA")
  bad3 = TF_VAR_foo_bar

  test_input = {
    input1  = dependency.test2.outputs.obj.foo
    input2  = dependency.test2.outputs.obj.bar
    input3  = dependency.test2.outputs.improper_mock.foo
    input4  = dependency.test2.outputs.improper_mock.bar
    input5  = dependency.test2.outputs.list[0].foo
    input6  = dependency.test2.outputs.list[1].bar
    input7  = dependency.test2.outputs.improper_mock2[0].foo
    input8  = dependency.test2.outputs.map["foo"].bar
    input9  = dependency.test2.outputs.map["baz"].bat
    input10 = dependency.test2.outputs.improper_mock3["foo"].bar
    input11 = dependency.test2.outputs.not_exist.foo
    input12 = dependency.test2.outputs.not_exist_list[0].foo
    input13 = dependency.test2.outputs.not_exist_map["foo"].bar
    input14 = dependency.test2.outputs.list_simple[0]
    input15 = dependency.test2.outputs.list_simple[0]
    input16 = dependency.test2.outputs.map_simple["foo"]
    input17 = dependency.test2.outputs.map_simple["bar"]
    input18 = dependency.test2.outputs.list_not_exist_simple[0]
    input19 = dependency.test2.outputs.map_not_exist_simple["foo"]
    input20 = dependency.test2.outputs.list_simple_existing[1]
    input21 = dependency.test2.outputs.list_existing[1].foo
    input22 = dependency.test2.outputs.map_existing["bar"]
    input23 = dependency.test2.outputs.bad
    input24 = dependency.test2.outputs.bad_list[0]
    input24 = dependency.test2.outputs.good_list_bad_prop[0].id
    input24 = dependency.test2.outputs.bad_map["test"]
    input25 = dependency.test2.outputs.bad_map["test"]["bar"].id
    input26 = [dependency.test2.outputs.map_existing["test"]["bar"].id]
    input27 = [dependency.test2.outputs.map_existing["test"][local.region].id]
    input28 = dependency.test2.outputs.map_existing["test"][local.region].id
    input29 = <<EOF
    {
      "key": "${dependency.stag-dep.outputs.retrieved_secrets.0}",
      "key2": "${dependency.stag-dep.outputs.retrieved_secrets.1}"
    }
EOF
  }

  hello_world_function_memory_size = 512
}
