provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_get_ec2_platforms      = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_launch_configuration" "lc_basic" {
  image_id      = "fake_ami"
  instance_type = "t2.medium"

  root_block_device {
    volume_size = 10
  }

  ebs_block_device {
    device_name = "xvdf"
    volume_size = 10
  }

  ebs_block_device {
    device_name = "xvdg"
    volume_type = "gp3"
    volume_size = 10
  }
}

resource "aws_autoscaling_group" "asg_lc_basic" {
  launch_configuration = aws_launch_configuration.lc_basic.id
  desired_capacity     = 2
  max_size             = 3
  min_size             = 1
}

resource "aws_launch_configuration" "lc_ebs_optimized" {
  image_id          = "fake_ami"
  instance_type     = "r3.xlarge"
  enable_monitoring = false
  ebs_optimized     = true
}

resource "aws_autoscaling_group" "asg_lc_ebs_optimized" {
  launch_configuration = aws_launch_configuration.lc_ebs_optimized.id
  desired_capacity     = 2
  max_size             = 3
  min_size             = 1
}

resource "aws_launch_configuration" "lc_tenancy_dedicated" {
  image_id          = "fake_ami"
  instance_type     = "m3.medium"
  placement_tenancy = "dedicated"
  enable_monitoring = false
}

resource "aws_autoscaling_group" "asg_lc_tenancy_dedicated" {
  launch_configuration = aws_launch_configuration.lc_tenancy_dedicated.id
  desired_capacity     = 2
  max_size             = 3
  min_size             = 1
}

resource "aws_launch_configuration" "lc_tenancy_host" {
  image_id          = "fake_ami"
  instance_type     = "m3.medium"
  placement_tenancy = "host"
  enable_monitoring = false
}

resource "aws_autoscaling_group" "asg_lc_tenancy_host" {
  launch_configuration = aws_launch_configuration.lc_tenancy_host.id
  desired_capacity     = 2
  max_size             = 3
  min_size             = 1
}

resource "aws_launch_configuration" "lc_cpu_credits_noUsage" {
  image_id          = "fake_ami"
  instance_type     = "t3.medium"
  enable_monitoring = false
}

resource "aws_autoscaling_group" "asg_lc_cpu_credits_noUsage" {
  launch_configuration = aws_launch_configuration.lc_cpu_credits_noUsage.id
  desired_capacity     = 2
  max_size             = 3
  min_size             = 1
}

resource "aws_launch_configuration" "lc_cpu_credits" {
  image_id          = "fake_ami"
  instance_type     = "t3.medium"
  enable_monitoring = false
}

resource "aws_autoscaling_group" "asg_lc_cpu_credits" {
  launch_configuration = aws_launch_configuration.lc_cpu_credits.id
  desired_capacity     = 2
  max_size             = 3
  min_size             = 1
}

resource "aws_launch_configuration" "lc_usage" {
  image_id      = "fake_ami"
  instance_type = "t2.medium"

  root_block_device {
    volume_size = 10
  }

  ebs_block_device {
    device_name = "xvdf"
    volume_size = 10
  }
}

resource "aws_autoscaling_group" "asg_lc_usage" {
  launch_configuration = aws_launch_configuration.lc_usage.id
  desired_capacity     = 2
  max_size             = 3
  min_size             = 1
}

resource "aws_launch_configuration" "lc_reserved" {
  image_id      = "fake_ami"
  instance_type = "t3.medium"
}

resource "aws_autoscaling_group" "asg_lc_reserved" {
  launch_configuration = aws_launch_configuration.lc_reserved.id
  desired_capacity     = 1
  max_size             = 1
  min_size             = 1
}

resource "aws_launch_configuration" "lc_windows" {
  image_id      = "fake_ami"
  instance_type = "t3.medium"
}

resource "aws_autoscaling_group" "asg_lc_windows" {
  launch_configuration = aws_launch_configuration.lc_windows.id
  desired_capacity     = 1
  max_size             = 1
  min_size             = 1
}
