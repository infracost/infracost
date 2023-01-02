provider "aws" {
  region = local.region
}

locals {
  name   = "example-ec2-complete"
  region = "eu-west-1"

  user_data = <<-EOT
  #!/bin/bash
  echo "Hello Terraform!"
  EOT

  tags = {
    Owner       = "user"
    Environment = "dev"
  }
}

################################################################################
# Supporting Resources
################################################################################

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 3.0"

  name = local.name
  cidr = "10.99.0.0/18"

  azs              = ["${local.region}a", "${local.region}b", "${local.region}c"]
  public_subnets   = ["10.99.0.0/24", "10.99.1.0/24", "10.99.2.0/24"]
  private_subnets  = ["10.99.3.0/24", "10.99.4.0/24", "10.99.5.0/24"]
  database_subnets = ["10.99.7.0/24", "10.99.8.0/24", "10.99.9.0/24"]

  tags = local.tags
}

data "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

module "security_group" {
  source  = "terraform-aws-modules/security-group/aws"
  version = "~> 4.0"

  name        = local.name
  description = "Security group for example usage with EC2 instance"
  vpc_id      = module.vpc.vpc_id

  ingress_cidr_blocks = ["0.0.0.0/0"]
  ingress_rules       = ["http-80-tcp", "all-icmp"]
  egress_rules        = ["all-all"]

  tags = local.tags
}

resource "aws_placement_group" "web" {
  name     = local.name
  strategy = "cluster"
}

resource "aws_kms_key" "this" {
}

resource "aws_network_interface" "this" {
  subnet_id = element(module.vpc.private_subnets, 0)
}

################################################################################
# EC2 Module
################################################################################

module "ec2_disabled" {
  source = "../../"

  create = false
}

module "ec2_complete" {
  source = "../../"

  name = local.name

  ami                         = data.aws_ami.amazon_linux.id
  instance_type               = "c5.4xlarge"
  availability_zone           = element(module.vpc.azs, 0)
  subnet_id                   = element(module.vpc.private_subnets, 0)
  vpc_security_group_ids      = [module.security_group.security_group_id]
  placement_group             = aws_placement_group.web.id
  associate_public_ip_address = true

  # only one of these can be enabled at a time
  hibernation = true
  # enclave_options_enabled = true

  user_data_base64 = base64encode(local.user_data)

  cpu_core_count       = 2 # default 4
  cpu_threads_per_core = 1 # default 2

  capacity_reservation_specification = {
    capacity_reservation_preference = "open"
  }

  enable_volume_tags = false
  root_block_device = [
    {
      encrypted   = true
      volume_type = "gp3"
      throughput  = 200
      volume_size = 50
      tags = {
        Name = "my-root-block"
      }
    },
  ]

  ebs_block_device = [
    {
      device_name = "/dev/sdf"
      volume_type = "gp3"
      volume_size = 5
      throughput  = 200
      encrypted   = true
      kms_key_id  = aws_kms_key.this.arn
    }
  ]

  tags = local.tags
}

module "ec2_network_interface" {
  source = "../../"

  name = "${local.name}-network-interface"

  ami           = data.aws_ami.amazon_linux.id
  instance_type = "c5.large"

  network_interface = [
    {
      device_index          = 0
      network_interface_id  = aws_network_interface.this.id
      delete_on_termination = false
    }
  ]

  tags = local.tags
}

module "ec2_metadata_options" {
  source = "../../"

  name = "${local.name}-metadata-options"

  ami                    = data.aws_ami.amazon_linux.id
  instance_type          = "c5.4xlarge"
  subnet_id              = element(module.vpc.private_subnets, 0)
  vpc_security_group_ids = [module.security_group.security_group_id]

  metadata_options = {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 8
    instance_metadata_tags      = "enabled"
  }

  tags = local.tags
}

module "ec2_t2_unlimited" {
  source = "../../"

  name = "${local.name}-t2-unlimited"

  ami                         = data.aws_ami.amazon_linux.id
  instance_type               = "t2.micro"
  cpu_credits                 = "unlimited"
  subnet_id                   = element(module.vpc.private_subnets, 0)
  vpc_security_group_ids      = [module.security_group.security_group_id]
  associate_public_ip_address = true

  tags = local.tags
}

module "ec2_t3_unlimited" {
  source = "../../"

  name = "${local.name}-t3-unlimited"

  ami                         = data.aws_ami.amazon_linux.id
  instance_type               = "t3.micro"
  cpu_credits                 = "unlimited"
  subnet_id                   = element(module.vpc.private_subnets, 0)
  vpc_security_group_ids      = [module.security_group.security_group_id]
  associate_public_ip_address = true

  tags = local.tags
}

################################################################################
# EC2 Module - multiple instances with `for_each`
################################################################################

locals {
  multiple_instances = {
    one = {
      instance_type     = "t3.micro"
      availability_zone = element(module.vpc.azs, 0)
      subnet_id         = element(module.vpc.private_subnets, 0)
      root_block_device = [
        {
          encrypted   = true
          volume_type = "gp3"
          throughput  = 200
          volume_size = 50
          tags = {
            Name = "my-root-block"
          }
        }
      ]
    }
    two = {
      instance_type     = "t3.small"
      availability_zone = element(module.vpc.azs, 1)
      subnet_id         = element(module.vpc.private_subnets, 1)
      root_block_device = [
        {
          encrypted   = true
          volume_type = "gp2"
          volume_size = 50
        }
      ]
    }
    three = {
      instance_type     = "t3.medium"
      availability_zone = element(module.vpc.azs, 2)
      subnet_id         = element(module.vpc.private_subnets, 2)
    }
  }
}

module "ec2_multiple" {
  source = "../../"

  for_each = local.multiple_instances

  name = "${local.name}-multi-${each.key}"

  ami                    = data.aws_ami.amazon_linux.id
  instance_type          = each.value.instance_type
  availability_zone      = each.value.availability_zone
  subnet_id              = each.value.subnet_id
  vpc_security_group_ids = [module.security_group.security_group_id]

  enable_volume_tags = false
  root_block_device  = lookup(each.value, "root_block_device", [])

  tags = local.tags
}

################################################################################
# EC2 Module - spot instance request
################################################################################

module "ec2_spot_instance" {
  source = "../../"

  name                 = "${local.name}-spot-instance"
  create_spot_instance = true

  ami                         = data.aws_ami.amazon_linux.id
  instance_type               = "c4.4xlarge"
  availability_zone           = element(module.vpc.azs, 0)
  subnet_id                   = element(module.vpc.private_subnets, 0)
  vpc_security_group_ids      = [module.security_group.security_group_id]
  placement_group             = aws_placement_group.web.id
  associate_public_ip_address = true

  # Spot request specific attributes
  spot_price                          = "0.60"
  spot_wait_for_fulfillment           = true
  spot_type                           = "persistent"
  spot_instance_interruption_behavior = "terminate"
  # End spot request specific attributes

  user_data_base64 = base64encode(local.user_data)

  cpu_core_count       = 2 # default 4
  cpu_threads_per_core = 1 # default 2

  capacity_reservation_specification = {
    capacity_reservation_preference = "open"
  }

  enable_volume_tags = false
  root_block_device = [
    {
      encrypted   = true
      volume_type = "gp3"
      throughput  = 200
      volume_size = 50
      tags = {
        Name = "my-root-block"
      }
    },
  ]

  ebs_block_device = [
    {
      device_name = "/dev/sdf"
      volume_type = "gp3"
      volume_size = 5
      throughput  = 200
      encrypted   = true
      # kms_key_id  = aws_kms_key.this.arn # you must grant the AWSServiceRoleForEC2Spot service-linked role access to any custom KMS keys
    }
  ]

  tags = local.tags
}
