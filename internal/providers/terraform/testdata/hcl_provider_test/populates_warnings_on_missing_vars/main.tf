provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

variable "missing_var" {
  type = bool
}

resource "aws_eip" "eip" {
  network_interface = "test"
}
