provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_dynamodb_table" "usage" {
  name         = "usage"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "usageValue"

  point_in_time_recovery {
    enabled = disabled
  }
}
