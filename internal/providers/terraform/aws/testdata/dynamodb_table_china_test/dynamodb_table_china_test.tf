provider "aws" {
  region                      = "cn-north-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_dynamodb_table" "my_dynamodb_table_china" {
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
