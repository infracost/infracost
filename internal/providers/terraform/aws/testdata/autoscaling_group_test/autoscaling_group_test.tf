provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
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

resource "aws_autoscaling_group" "asg_lc_min_size" {
  launch_configuration = aws_launch_configuration.lc_basic.id
  max_size             = 3
  min_size             = 2
}

resource "aws_autoscaling_group" "asg_lc_min_size_name_ref" {
  launch_configuration = aws_launch_configuration.lc_basic.name
  max_size             = 3
  min_size             = 2
}

resource "aws_autoscaling_group" "asg_lc_min_size_zero" {
  launch_configuration = aws_launch_configuration.lc_basic.id
  max_size             = 3
  min_size             = 0
}

resource "aws_autoscaling_group" "asg_lc_desired_capacity_zero" {
  launch_configuration = aws_launch_configuration.lc_basic.id
  desired_capacity     = 0
  max_size             = 3
  min_size             = 0
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

resource "aws_launch_template" "lt_basic" {
  image_id      = "fake_ami"
  instance_type = "t2.medium"

  block_device_mappings {
    device_name = "xvdf"
    ebs {
      volume_size = 10
    }
  }

  block_device_mappings {
    device_name = "xvfa"
    ebs {
      volume_size = 20
      volume_type = "io1"
      iops        = 200
    }
  }
}

resource "aws_autoscaling_group" "asg_lt_basic" {
  launch_template {
    id = aws_launch_template.lt_basic.id
  }
  desired_capacity = 2
  max_size         = 3
  min_size         = 1
}

resource "aws_autoscaling_group" "asg_lt_min_size" {
  launch_template {
    id = aws_launch_template.lt_basic.id
  }
  max_size = 3
  min_size = 2
}

resource "aws_autoscaling_group" "asg_lt_min_size_zero" {
  launch_template {
    id = aws_launch_template.lt_basic.id
  }
  max_size = 3
  min_size = 0
}

resource "aws_autoscaling_group" "asg_lt_desired_capacity_zero" {
  launch_template {
    id = aws_launch_template.lt_basic.id
  }
  desired_capacity = 0
  max_size         = 3
  min_size         = 0
}

resource "aws_launch_template" "lt_tenancy_dedicated" {
  image_id      = "fake_ami"
  instance_type = "m3.medium"
  placement {
    tenancy = "dedicated"
  }
}

resource "aws_autoscaling_group" "asg_lt_tenancy_dedicated" {
  launch_template {
    id = aws_launch_template.lt_tenancy_dedicated.id
  }
  desired_capacity = 2
  max_size         = 3
  min_size         = 1
}

resource "aws_launch_template" "lt_tenancy_host" {
  image_id      = "fake_ami"
  instance_type = "m3.medium"
  placement {
    tenancy = "host"
  }
}

resource "aws_autoscaling_group" "asg_lt_tenancy_host" {
  launch_template {
    id = aws_launch_template.lt_tenancy_host.id
  }
  desired_capacity = 2
  max_size         = 3
  min_size         = 1
}

resource "aws_launch_template" "lt_ebs_optimized" {
  image_id      = "fake_ami"
  instance_type = "r3.xlarge"
  ebs_optimized = true
}

resource "aws_autoscaling_group" "asg_lt_ebs_optimized" {
  launch_template {
    id = aws_launch_template.lt_ebs_optimized.id
  }
  desired_capacity = 2
  max_size         = 3
  min_size         = 1
}

resource "aws_launch_template" "lt_elastic_inference_accelerator" {
  image_id      = "fake_ami"
  instance_type = "t2.medium"
  elastic_inference_accelerator {
    type = "eia2.medium"
  }
}

resource "aws_autoscaling_group" "asg_lt_elastic_inference_accelerator" {
  launch_template {
    id = aws_launch_template.lt_elastic_inference_accelerator.id
  }
  desired_capacity = 2
  max_size         = 3
  min_size         = 1
}

resource "aws_launch_template" "lt_monitoring" {
  image_id      = "fake_ami"
  instance_type = "t2.medium"
  monitoring {
    enabled = true
  }
}

resource "aws_autoscaling_group" "asg_lt_monitoring" {
  launch_template {
    id = aws_launch_template.lt_monitoring.id
  }
  desired_capacity = 2
  max_size         = 3
  min_size         = 1
}

resource "aws_launch_template" "lt_cpu_credits_noUsage" {
  image_id      = "fake_ami"
  instance_type = "t3.large"
}

resource "aws_autoscaling_group" "asg_lt_cpu_credits_noUsage" {
  launch_template {
    id = aws_launch_template.lt_cpu_credits_noUsage.id
  }
  desired_capacity = 2
  max_size         = 3
  min_size         = 1
}

resource "aws_launch_template" "lt_cpu_credits" {
  image_id      = "fake_ami"
  instance_type = "t3.large"
}

resource "aws_autoscaling_group" "asg_lt_cpu_credits" {
  launch_template {
    id = aws_launch_template.lt_cpu_credits.id
  }
  desired_capacity = 2
  max_size         = 3
  min_size         = 1
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

resource "aws_autoscaling_group" "asg_lt_usage" {
  launch_template {
    id = aws_launch_template.lt_usage.id
  }
  desired_capacity = 2
  max_size         = 3
  min_size         = 1
}

resource "aws_launch_template" "lt_mixed_instance_basic" {
  image_id      = "fake_ami"
  instance_type = "t2.medium"
}

resource "aws_autoscaling_group" "asg_mixed_instance_basic" {
  desired_capacity = 6
  max_size         = 10
  min_size         = 1

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.lt_mixed_instance_basic.id
      }

      override {
        instance_type     = "t2.large"
        weighted_capacity = "2"
      }

      override {
        instance_type     = "t2.xlarge"
        weighted_capacity = "4"
      }
    }

    instances_distribution {
      on_demand_base_capacity                  = 1
      on_demand_percentage_above_base_capacity = 100
    }
  }
}

resource "aws_launch_template" "lt_mixed_instance_dynamic" {
  image_id      = "fake_ami"
  instance_type = "t2.medium"
}

resource "aws_autoscaling_group" "asg_mixed_instance_dynamic" {
  desired_capacity = 3
  max_size         = 5
  min_size         = 1

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.lt_mixed_instance_dynamic.id
      }

      dynamic "override" {
        for_each = ["t2.large", "t2.xlarge"]

        content {
          instance_type = override.value
        }
      }
    }

    instances_distribution {
      on_demand_base_capacity                  = 1
      on_demand_percentage_above_base_capacity = 100
    }
  }
}

module "asg-lt" {
  source                 = "terraform-aws-modules/autoscaling/aws"
  version                = "~> 5"
  name                   = "asg"
  create_launch_template = true
  launch_template_name   = "lt"
  image_id               = "ami-0ff8a91507f77f867"
  instance_type          = "t3.micro"
  min_size               = 0
  max_size               = 2
  desired_capacity       = 1
  block_device_mappings = [
    {
      device_name = "/dev/xvdf"
      ebs = {
        volume_size = 10
      }
    }
  ]
}


locals {
  instance_types = ["t2.micro", "t2.medium"]
}

resource "aws_autoscaling_group" "test_count" {
  count                = 2
  desired_capacity     = 2
  max_size             = 3
  min_size             = 1
  launch_configuration = aws_launch_configuration.test_count.*.id[count.index]
}

resource "aws_launch_configuration" "test_count" {
  count         = 2
  image_id      = "fake_ami"
  instance_type = local.instance_types[count.index]
}
