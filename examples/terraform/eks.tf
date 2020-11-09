resource "aws_eks_cluster" "example" {
  name     = "example"
  role_arn = "arn:aws:iam::123456789012:role/Example"

  vpc_config {
    subnet_ids = ["subnet_id"]
  }
}

resource "aws_eks_fargate_profile" "example" {
  cluster_name           = aws_eks_cluster.example.name
  fargate_profile_name   = "example"
  pod_execution_role_arn = "arn:aws:iam::123456789012:role/Example"
  subnet_ids             = ["subnet_id"]

  selector {
    namespace = "example"
  }
}

resource "aws_eks_node_group" "example" {
  cluster_name    = "test aws_eks_node_group"
  node_group_name = "example"
  instance_types  = ["t2.medium"]
  node_role_arn   = "node_role_arn"
  subnet_ids      = ["subnet_id"]

  scaling_config {
    desired_size = 1
    max_size     = 1
    min_size     = 1
  }
}

resource "aws_eks_node_group" "example_with_launch_template" {
  cluster_name    = "test aws_eks_node_group"
  node_group_name = "example"
  node_role_arn   = "node_role_arn"
  subnet_ids      = ["subnet_id"]

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
    cpu_credits = "standard"
  }

  disable_api_termination = true

  ebs_optimized = true

  elastic_gpu_specifications {
    type = "test"
  }

  elastic_inference_accelerator {
    type = "eia1.medium"
  }

  iam_instance_profile {
    name = "test"
  }

  image_id = "ami-test"

  instance_initiated_shutdown_behavior = "terminate"

  instance_market_options {
    market_type = "spot"
  }

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

  monitoring {
    enabled = true
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
