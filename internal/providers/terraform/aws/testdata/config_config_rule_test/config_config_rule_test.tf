provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_config_config_rule" "my_config" {
  name = "example"

  source {
    owner             = "AWS"
    source_identifier = ""
  }
}

resource "aws_config_config_rule" "my_config_withUsage" {
  name = "example"

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_VERSIONING_ENABLED"
  }
}
