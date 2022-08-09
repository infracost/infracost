provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_eip" "invalid_eip" {
  network_interface = "test"

  network_interface {
    this_is_wrong = "test"
  }
}
