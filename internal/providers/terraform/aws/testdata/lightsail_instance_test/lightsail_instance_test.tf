provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_lightsail_instance" "linux1" {
  name              = "linux1"
  availability_zone = "us-east-1a"
  blueprint_id      = "centos_7_1901_01"
  bundle_id         = "xlarge_2_0"
}

resource "aws_lightsail_instance" "win1" {
  name              = "win1"
  availability_zone = "us-east-1a"
  blueprint_id      = "windows_2019"
  bundle_id         = "small_win_2_0"
}
