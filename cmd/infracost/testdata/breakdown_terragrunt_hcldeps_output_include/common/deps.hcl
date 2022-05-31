dependency "test" {
  config_path = "../prod"
  mock_outputs = {
    aws_instance_type = "t2.micro"
  }
}

dependency "test2" {
  config_path = "../prod3"
  mock_outputs = {
    block_iops = 800
  }
}
