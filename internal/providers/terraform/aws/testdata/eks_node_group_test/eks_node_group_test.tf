provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

locals {
  launch_template_name = aws_launch_template.prefix.name
}

resource "aws_eks_node_group" "example" {
  cluster_name    = "test_aws_eks_node_group"
  node_group_name = "example"
  node_role_arn   = "node_role_arn"
  subnet_ids      = ["subnet_id"]

  scaling_config {
    desired_size = 1
    max_size     = 1
    min_size     = 1
  }
}

resource "aws_eks_node_group" "example_defaultCpuCredits" {
  cluster_name    = "test_aws_eks_node_group"
  node_group_name = "example"
  node_role_arn   = "node_role_arn"
  subnet_ids      = ["subnet_id"]

  scaling_config {
    desired_size = 3
    max_size     = 3
    min_size     = 1
  }
}

resource "aws_eks_node_group" "example_with_launch_template" {
  cluster_name    = "test_aws_eks_node_group"
  node_group_name = "example"
  node_role_arn   = "node_role_arn"
  subnet_ids      = ["subnet_id"]

  instance_types = ["t3.medium"]

  scaling_config {
    desired_size = 3
    max_size     = 1
    min_size     = 1
  }

  launch_template {
    id      = aws_launch_template.foo.id
    version = "default_version"
  }
}

resource "aws_eks_node_group" "example_with_launch_template_instance_types_default" {
  cluster_name    = "test_aws_eks_node_group"
  node_group_name = "example"
  node_role_arn   = "node_role_arn"
  subnet_ids      = ["subnet_id"]

  scaling_config {
    desired_size = 3
    max_size     = 1
    min_size     = 1
  }

  launch_template {
    id      = aws_launch_template.bar.id
    version = "default_version"
  }
}

resource "aws_launch_template" "bar" {
  name = "foo"

  block_device_mappings {
    device_name = "/dev/sda1"

    ebs {
      volume_size = 20
    }
  }

  capacity_reservation_specification {
    capacity_reservation_preference = "open"
  }

  cpu_options {
    core_count       = 4
    threads_per_core = 2
  }

  credit_specification {
    cpu_credits = "unlimited"
  }

  disable_api_termination = true

  ebs_optimized = true

  iam_instance_profile {
    name = "test"
  }

  image_id = "ami-test"

  instance_initiated_shutdown_behavior = "terminate"

  kernel_id = "test"

  key_name = "test"

  license_specification {
    license_configuration_arn = "arn:aws:license-manager:eu-west-1:123456789012:license-configuration:lic-0123456789abcdef0123456789abcdef"
  }

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 1
  }

  network_interfaces {
    associate_public_ip_address = true
  }

  placement {
    availability_zone = "us-west-2a"
  }

  ram_disk_id = "test"

  vpc_security_group_ids = ["example"]
}

resource "aws_launch_template" "foo" {
  name = "foo"

  block_device_mappings {
    device_name = "/dev/sda1"

    ebs {
      volume_size = 20
    }
  }

  capacity_reservation_specification {
    capacity_reservation_preference = "open"
  }

  cpu_options {
    core_count       = 4
    threads_per_core = 2
  }

  credit_specification {
    cpu_credits = "unlimited"
  }

  disable_api_termination = true

  ebs_optimized = true

  iam_instance_profile {
    name = "test"
  }

  image_id = "ami-test"

  instance_initiated_shutdown_behavior = "terminate"

  kernel_id = "test"

  key_name = "test"

  license_specification {
    license_configuration_arn = "arn:aws:license-manager:eu-west-1:123456789012:license-configuration:lic-0123456789abcdef0123456789abcdef"
  }

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 1
  }

  network_interfaces {
    associate_public_ip_address = true
  }

  placement {
    availability_zone = "us-west-2a"
  }

  ram_disk_id = "test"

  vpc_security_group_ids = ["example"]
}

resource "aws_eks_node_group" "example2" {
  cluster_name    = "test_aws_eks_node_group"
  node_group_name = "example"
  instance_types  = ["t2.medium"]
  node_role_arn   = "node_role_arn"
  disk_size       = 30
  subnet_ids      = ["subnet_id"]

  scaling_config {
    desired_size = 1
    max_size     = 1
    min_size     = 1
  }
}

resource "aws_eks_node_group" "example_with_launch_template_2" {
  cluster_name    = "test_aws_eks_node_group"
  node_group_name = "example"
  node_role_arn   = "node_role_arn"
  subnet_ids      = ["subnet_id"]

  scaling_config {
    desired_size = 3
    max_size     = 1
    min_size     = 1
  }

  launch_template {
    id      = aws_launch_template.foo2.id
    version = "default_version"
  }
}

resource "aws_launch_template" "foo2" {
  name = "foo2"

  block_device_mappings {
    device_name = "/dev/sda1"

    ebs {
      volume_size = 20
    }
  }

  capacity_reservation_specification {
    capacity_reservation_preference = "open"
  }

  cpu_options {
    core_count       = 4
    threads_per_core = 2
  }

  credit_specification {
    cpu_credits = "standard"
  }

  disable_api_termination = true

  ebs_optimized = true

  iam_instance_profile {
    name = "test"
  }

  image_id = "ami-test"

  instance_initiated_shutdown_behavior = "terminate"

  instance_type = "m5.xlarge"

  kernel_id = "test"

  key_name = "test"

  license_specification {
    license_configuration_arn = "arn:aws:license-manager:eu-west-1:123456789012:license-configuration:lic-0123456789abcdef0123456789abcdef"
  }

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 1
  }

  network_interfaces {
    associate_public_ip_address = true
  }

  placement {
    availability_zone = "us-west-2a"
  }

  ram_disk_id = "test"

  vpc_security_group_ids = ["example"]

}

resource "aws_eks_node_group" "example_with_launch_template_3" {
  cluster_name    = "test_aws_eks_node_group"
  node_group_name = "example"
  node_role_arn   = "node_role_arn"
  subnet_ids      = ["subnet_id"]

  instance_types = ["m5.large"]

  scaling_config {
    desired_size = 3
    max_size     = 1
    min_size     = 1
  }

  launch_template {
    name    = aws_launch_template.foo3.name
    version = "default_version"
  }
}

resource "aws_launch_template" "prefix" {
  name        = null
  name_prefix = "prefix-"

  block_device_mappings {
    device_name = "/dev/sda1"

    ebs {
      volume_type = "gp3"
      volume_size = 100
    }
  }

  capacity_reservation_specification {
    capacity_reservation_preference = "open"
  }

  cpu_options {
    core_count       = 4
    threads_per_core = 2
  }

  credit_specification {
    cpu_credits = "standard"
  }

  disable_api_termination = true

  ebs_optimized = true

  iam_instance_profile {
    name = "test"
  }

  image_id = "ami-test"

  instance_initiated_shutdown_behavior = "terminate"

  instance_type = "m5.xlarge"

  kernel_id = "test"

  key_name = "test"

  license_specification {
    license_configuration_arn = "arn:aws:license-manager:eu-west-1:123456789012:license-configuration:lic-0123456789abcdef0123456789abcdef"
  }

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 1
  }

  network_interfaces {
    associate_public_ip_address = true
  }

  placement {
    availability_zone = "us-west-2a"
  }

  ram_disk_id = "test"

  vpc_security_group_ids = ["example"]

}

resource "aws_eks_node_group" "example_with_launch_template_with_prefix" {
  cluster_name    = "test_aws_eks_node_group"
  node_group_name = "example"
  node_role_arn   = "node_role_arn"
  subnet_ids      = ["subnet_id"]

  instance_types = ["m5.large"]

  scaling_config {
    desired_size = 3
    max_size     = 1
    min_size     = 1
  }

  launch_template {
    name    = local.launch_template_name
    version = "default_version"
  }
}

resource "aws_launch_template" "foo3" {
  name = "foo3"

  block_device_mappings {
    device_name = "/dev/sda1"

    ebs {
      volume_size = 20
    }
  }

  capacity_reservation_specification {
    capacity_reservation_preference = "open"
  }

  cpu_options {
    core_count       = 4
    threads_per_core = 2
  }

  credit_specification {
    cpu_credits = "standard"
  }

  disable_api_termination = true

  ebs_optimized = true

  iam_instance_profile {
    name = "test"
  }

  image_id = "ami-test"

  instance_initiated_shutdown_behavior = "terminate"

  instance_type = "m5.xlarge"

  kernel_id = "test"

  key_name = "test"

  license_specification {
    license_configuration_arn = "arn:aws:license-manager:eu-west-1:123456789012:license-configuration:lic-0123456789abcdef0123456789abcdef"
  }

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 1
  }

  network_interfaces {
    associate_public_ip_address = true
  }

  placement {
    availability_zone = "us-west-2a"
  }

  ram_disk_id = "test"

  vpc_security_group_ids = ["example"]

}

resource "aws_eks_node_group" "reserved" {
  cluster_name    = "test_aws_eks_node_group"
  node_group_name = "example"
  node_role_arn   = "node_role_arn"
  subnet_ids      = ["subnet_id"]

  scaling_config {
    desired_size = 1
    max_size     = 1
    min_size     = 1
  }
}

resource "aws_eks_node_group" "windows" {
  cluster_name    = "test_aws_eks_node_group"
  node_group_name = "example"
  node_role_arn   = "node_role_arn"
  subnet_ids      = ["subnet_id"]

  scaling_config {
    desired_size = 1
    max_size     = 1
    min_size     = 1
  }
}

resource "aws_eks_node_group" "usage" {
  cluster_name    = "test_aws_eks_node_group"
  node_group_name = "example"
  node_role_arn   = "node_role_arn"
  subnet_ids      = ["subnet_id"]

  scaling_config {
    desired_size = 1
    max_size     = 1
    min_size     = 1
  }
}

resource "aws_launch_template" "lt_usage" {
  image_id      = "fake_ami"
  instance_type = "t2.medium"

  block_device_mappings {
    device_name = "xvdf"
    ebs {
      volume_size = 10
    }
  }
}

resource "aws_eks_node_group" "lt_usage" {
  cluster_name    = "test_aws_eks_node_group"
  node_group_name = "example"
  node_role_arn   = "node_role_arn"
  subnet_ids      = ["subnet_id"]

  scaling_config {
    desired_size = 1
    max_size     = 1
    min_size     = 1
  }

  launch_template {
    id      = aws_launch_template.lt_usage.id
    version = "default_version"
  }
}

