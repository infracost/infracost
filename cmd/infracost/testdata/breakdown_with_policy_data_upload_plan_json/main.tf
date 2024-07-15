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

resource "aws_instance" "web_app" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge"
  volume_tags = {
    "baz" = "bat"
  }

  root_block_device {
    volume_size = 51
  }

  ebs_block_device {
    device_name = "my_data"
    volume_type = "gp2"
    volume_size = 1000
    iops        = 800
  }

  tags = {
    "foo" = "bar"
  }
}

resource "aws_dynamodb_table" "autoscale_dynamodb_table" {
  name           = "GameScores"
  billing_mode   = "PROVISIONED"
  read_capacity  = 30
  write_capacity = 20
  hash_key       = "UserId"
  range_key      = "GameTitle"

  attribute {
    name = "UserId"
    type = "S"
  }

  attribute {
    name = "GameTitle"
    type = "S"
  }

  point_in_time_recovery {
    enabled = true
  }
}

resource "aws_appautoscaling_target" "autoscale_dynamodb_table_write_target" {
  max_capacity       = 99
  min_capacity       = 6
  resource_id        = "table/${aws_dynamodb_table.autoscale_dynamodb_table.name}"
  scalable_dimension = "dynamodb:table:WriteCapacityUnits"
  service_namespace  = "dynamodb"
}

resource "aws_appautoscaling_target" "autoscale_dynamodb_table_read_target" {
  max_capacity       = 100
  min_capacity       = 5
  resource_id        = "table/LiteralTableRef"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  service_namespace  = "dynamodb"
}
