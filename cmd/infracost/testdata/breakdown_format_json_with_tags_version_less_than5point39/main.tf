provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"

  default_tags {
    tags = {
      DefaultNotOverride = "defaultnotoverride"
      DefaultOverride    = "defaultoverride"
    }
  }
}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
}

resource "aws_sns_topic_subscription" "sns_topic_noTags" {
  endpoint  = "some-dummy-endpoint"
  protocol  = "http"
  topic_arn = "arn:aws:sns:us-east-1:123456789123:sns-topic-arn"
}

resource "aws_sns_topic_subscription" "sns_topic_withTags" {
  endpoint  = "some-dummy-endpoint"
  protocol  = "http"
  topic_arn = "arn:aws:sns:us-east-1:123456789123:sns-topic-arn"
  tags = {
    DefaultOverride = "sns-def"
    ResourceTag     = "sns-ghi"
  }
}

resource "aws_sqs_queue" "sqs_noTags" {
  name = "sqs_noTags"
}

resource "aws_sqs_queue" "sqs_withTags" {
  name = "sqs_withTags"

  tags = {
    DefaultOverride = "sqs-def"
    ResourceTag     = "sqs-hi"
  }
}

resource "aws_autoscaling_group" "bar" {
  name                      = "foobar3-terraform-test"
  max_size                  = 5
  min_size                  = 2
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 4
  force_delete              = true

  tag {
    key                 = "foo"
    value               = "bar"
    propagate_at_launch = true
  }

  tag {
    key                 = "lorem"
    value               = "ipsum"
    propagate_at_launch = false
  }
}

resource "aws_instance" "web_app" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge"
  volume_tags = {
    "baz"           = "bat"
    DefaultOverride = "volume_tag_overwritten"
  }

  root_block_device {
    volume_size = 50
  }

  ebs_block_device {
    device_name = "my_data"
    volume_type = "io1"
    volume_size = 1000
    iops        = 800
  }

  tags = {
    "foo" = "bar"
  }
}

resource "aws_instance" "launch_instance" {
  volume_tags = {
    "flip" = "flop"
  }
  launch_template {
    id = aws_launch_template.example.id
  }

  tags = {
    "bat"  = "ball"
    "fizz" = "buzz"
  }
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

  tag_specifications {
    resource_type = "instance"
    tags = {
      "fizz" = "buzz"
    }
  }

  tag_specifications {
    resource_type = "volume"
    tags = {
      "flip" = "flop"
    }
  }

  tag_specifications {
    resource_type = "capacity-reservation"
    tags = {
      "no" = "cap"
    }
  }
}

resource "aws_autoscaling_group" "launch_template" {
  name                      = "foobar3-terraform-test"
  max_size                  = 5
  min_size                  = 2
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 4
  force_delete              = true

  launch_template {
    id = aws_launch_template.example.id
  }

  tag {
    key                 = "foo"
    value               = "bar"
    propagate_at_launch = true
  }

  tag {
    key                 = "lorem"
    value               = "ipsum"
    propagate_at_launch = false
  }
}

resource "aws_instance" "web_app_block_tags" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge"

  root_block_device {
    volume_size = 50
    tags = {
      "foo" = "rbd"
    }
  }

  ebs_block_device {
    device_name = "my_data"
    volume_type = "io1"
    volume_size = 1000
    iops        = 800

    tags = {
      "foo" = "ebd"
    }
  }
}
