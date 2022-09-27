provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

variable "not_provided" {
  type = number
}

resource "aws_eip" "eip" {
  count = var.not_provided
}
