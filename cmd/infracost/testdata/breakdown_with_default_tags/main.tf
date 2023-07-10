provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"

  default_tags {
    tags = {
      Environment = "Test"
      Owner       = "TFProviders"
      Project     = "TestProject"
      SomeBool    = true
      SomeNumber  = 1
      SomeFloat   = 1.1
    }
  }
}



resource "aws_lambda_function" "hello_world" {
  function_name = "hello_world"
  role          = "arn:aws:lambda:us-east-1:aws:resource-id"
  handler       = "exports.test"
  runtime       = "nodejs12.x"
  filename      = "function.zip"
  tags = {
    Project   = "LambdaTestProject"
    lambdaTag = "hello"
  }
}

resource "aws_launch_configuration" "lc" {
  image_id      = "fake_ami"
  instance_type = "t3.medium"
}

variable "default_tags" {
  default = [
    {
      key                 = "Environment"
      value               = "Test-var"
      propagate_at_launch = true
    },
    {
      key                 = "Owner"
      value               = "TFProviders-var"
      propagate_at_launch = true
    },
    {
      key                 = "Project"
      value               = "TestProject-var"
      propagate_at_launch = true
    },
    {
      key                 = "SomeDynamicBool"
      value               = true
      propagate_at_launch = true
    },
    {
      key                 = "SomeDynamicNumber"
      value               = 1
      propagate_at_launch = true
    },
    {
      key                 = "SomeDynamicFloat"
      value               = 1.1
      propagate_at_launch = true
    },
  ]
  description = "Default Tags for Auto Scaling Group"
}

resource "aws_autoscaling_group" "asg" {
  launch_configuration = aws_launch_configuration.lc.id
  max_size             = 1
  min_size             = 1

  tag {
    key                 = "Project"
    value               = "ASGTestProject"
    propagate_at_launch = true
  }

  tag {
    key                 = "asgTag"
    value               = "locasgtag"
    propagate_at_launch = false
  }

  dynamic "tag" {
    for_each = var.default_tags
    content {
      key                 = tag.value.key
      propagate_at_launch = tag.value.propagate_at_launch
      value               = tag.value.value
    }
  }
}
