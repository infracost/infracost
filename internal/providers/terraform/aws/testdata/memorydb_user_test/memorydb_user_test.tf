provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_memorydb_user" "example" {
  user_name     = "example-user"
  access_string = "on ~* &* +@all"
  
  authentication_mode {
    type      = "password"
    passwords = ["password123456789"]
  }
}
