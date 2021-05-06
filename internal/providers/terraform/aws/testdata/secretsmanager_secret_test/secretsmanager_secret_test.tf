provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_secretsmanager_secret" "secret" {
  name = "my-test-secret"
}

resource "aws_secretsmanager_secret" "secret_withUsage" {
  name = "my-test-secret"
}