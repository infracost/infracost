provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_instance" "instance1" {
  ami           = "fake_ami"
  instance_type = "m3.medium"

  root_block_device {
    volume_size = 10
  }

  ebs_block_device {
    device_name = "xvdf"
    volume_size = 10
  }

  ebs_block_device {
    device_name = "xvdg"
    volume_type = "standard"
    volume_size = 20
  }

  ebs_block_device {
    device_name = "xvdh"
    volume_type = "sc1"
    volume_size = 30
  }

  ebs_block_device {
    device_name = "xvdi"
    volume_type = "io1"
    volume_size = 40
    iops        = 1000
  }

  ebs_block_device {
    device_name = "xvdj"
    volume_type = "gp3"
    volume_size = 20
  }
}

resource "aws_instance" "instance1_ebsOptimized" {
  ami           = "fake_ami"
  instance_type = "m3.large"
  ebs_optimized = true
}

resource "aws_instance" "instance2_ebsOptimized" {
  ami           = "fake_ami"
  instance_type = "r3.xlarge"
  ebs_optimized = true
}

resource "aws_instance" "instance1_hostTenancy" {
  ami           = "fake_ami"
  instance_type = "m3.medium"
  tenancy       = "host"
}

resource "aws_instance" "t3_default_cpuCredits" {
  ami           = "fake_ami"
  instance_type = "t3.medium"
}

resource "aws_instance" "t3_unlimited_cpuCredits" {
  ami           = "fake_ami"
  instance_type = "t3.medium"
  credit_specification {
    cpu_credits = "unlimited"
  }
}

resource "aws_instance" "t3_standard_cpuCredits" {
  ami           = "fake_ami"
  instance_type = "t3.medium"
  credit_specification {
    cpu_credits = "standard"
  }
}

resource "aws_instance" "t2_default_cpuCredits" {
  ami           = "fake_ami"
  instance_type = "t2.medium"
}

resource "aws_instance" "t2_unlimited_cpuCredits" {
  ami           = "fake_ami"
  instance_type = "t2.medium"
  credit_specification {
    cpu_credits = "unlimited"
  }
}

resource "aws_instance" "t2_standard_cpuCredits" {
  ami           = "fake_ami"
  instance_type = "t2.medium"
  credit_specification {
    cpu_credits = "standard"
  }
}

resource "aws_instance" "instance1_detailedMonitoring" {
  ami           = "fake_ami"
  instance_type = "m3.large"
  ebs_optimized = true
  monitoring    = true
}

resource "aws_instance" "std_1yr_no_upfront" {
  ami           = "fake_ami"
  instance_type = "t3.medium"
}
resource "aws_instance" "std_3yr_no_upfront" {
  ami           = "fake_ami"
  instance_type = "t3.medium"
}
resource "aws_instance" "std_1yr_partial_upfront" {
  ami           = "fake_ami"
  instance_type = "t3.medium"
}
resource "aws_instance" "std_3yr_partial_upfront" {
  ami           = "fake_ami"
  instance_type = "t3.medium"
}
resource "aws_instance" "std_1yr_all_upfront" {
  ami           = "fake_ami"
  instance_type = "t3.medium"
}
resource "aws_instance" "std_3yr_all_upfront" {
  ami           = "fake_ami"
  instance_type = "t3.medium"
}

resource "aws_instance" "cnvr_1yr_no_upfront" {
  ami           = "fake_ami"
  instance_type = "t3.medium"
}
resource "aws_instance" "cnvr_3yr_no_upfront" {
  ami           = "fake_ami"
  instance_type = "t3.medium"
}
resource "aws_instance" "cnvr_1yr_partial_upfront" {
  ami           = "fake_ami"
  instance_type = "t3.medium"
}
resource "aws_instance" "cnvr_3yr_partial_upfront" {
  ami           = "fake_ami"
  instance_type = "t3.medium"
}
resource "aws_instance" "cnvr_1yr_all_upfront" {
  ami           = "fake_ami"
  instance_type = "t3.medium"
}
resource "aws_instance" "cnvr_3yr_all_upfront" {
  ami           = "fake_ami"
  instance_type = "t3.medium"
}

resource "aws_launch_template" "example" {
  name = "example-lt"

  image_id      = "fake_ami"
  instance_type = "t3.medium"
  ebs_optimized = true

  monitoring {
    enabled = true
  }

  credit_specification {
    cpu_credits = "unlimited"
  }

  placement {
    tenancy = "dedicated"
  }

  block_device_mappings {
    device_name = "xvdc"

    ebs {
      volume_type = "io1"
      volume_size = 10
      iops        = 100
    }
  }
}

resource "aws_instance" "instance_withLaunchTemplateById" {
  launch_template {
    id = aws_launch_template.example.id
  }

  ebs_block_device {
    device_name = "xvdb"
    volume_type = "io1"
    volume_size = 20
    iops        = 200
  }
}

resource "aws_instance" "instance_withLaunchTemplateByName" {
  launch_template {
    name = aws_launch_template.example.name
  }
}

resource "aws_instance" "instance_withLaunchTemplateOverride" {
  ami           = "overriden-fake_ami"
  instance_type = "t3.large"
  ebs_optimized = false

  monitoring = false
  tenancy    = "default"

  credit_specification {
    cpu_credits = "standard"
  }

  launch_template {
    id = aws_launch_template.example.id
  }

  ebs_block_device {
    device_name = "xvdc"

    volume_size = 20
  }
}

resource "aws_instance" "instance_withMonthlyHours" {
  ami           = "fake_ami"
  instance_type = "t3.medium"
}

resource "aws_instance" "instance_ebsOptimized_withMonthlyHours" {
  ami           = "fake_ami"
  instance_type = "r3.xlarge"
  ebs_optimized = true
}

resource "aws_instance" "instance_detailedMonitoring_withMonthlyHours" {
  ami           = "fake_ami"
  instance_type = "m3.large"
  monitoring    = true
}

resource "aws_ec2_host" "mac" {
  instance_type     = "mac1.metal"
  availability_zone = "us-east-2a"
}

resource "aws_instance" "with_host" {
  ami           = "fake_ami"
  instance_type = "fake" # TF requires it
  host_id       = aws_ec2_host.mac.id

  root_block_device {
    volume_size = 50
  }
}
