provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_transfer_server" "default_no_protocols" {
  tags = {
    Name = "No protocols"
  }
}

resource "aws_transfer_server" "multiple_protocols_with_usage" {
  protocols = ["SFTP", "FTPS", "FTP"]

  tags = {
    Name = "Multiple protocols with usage"
  }
}
