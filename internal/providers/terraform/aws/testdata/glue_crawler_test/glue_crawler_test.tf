provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_glue_crawler" "no_usage" {
  database_name = "test"
  name          = "example"
  role          = "arn:aws:glue:us-east-1:123456789012:resource-id"

  dynamodb_target {
    path = "table-name"
  }
}

resource "aws_glue_crawler" "with_usage" {
  database_name = "test"
  name          = "example"
  role          = "arn:aws:glue:us-east-1:123456789012:resource-id"

  dynamodb_target {
    path = "table-name"
  }
}
