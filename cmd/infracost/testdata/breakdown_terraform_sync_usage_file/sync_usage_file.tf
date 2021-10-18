provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_nat_gateway" "example_each" {
  subnet_id = "subnet"

  for_each = toset(["assets", "media"])
  tags = {
    Name = "${each.key}_nat"
  }
}

resource "aws_nat_gateway" "example_count" {
  subnet_id = "subnet"
  count     = 2
}

resource "aws_cloudwatch_log_group" "production_logs" {
  for_each = toset(["assets", "media"])
  name     = "${each.key}_log"
}
