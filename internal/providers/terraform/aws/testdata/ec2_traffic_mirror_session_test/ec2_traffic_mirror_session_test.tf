provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_ec2_traffic_mirror_session" "session" {
  description              = "traffic mirror session"
  network_interface_id     = "eni-1234567"
  traffic_mirror_filter_id = "a-traffic-filter-id"
  traffic_mirror_target_id = "a-traffic-target-id"
  session_number           = "1"
}
