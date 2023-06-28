provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

# Add example resources for NetworkfirewallFirewall below

resource "aws_networkfirewall_firewall" "networkfirewall_firewall" {
  name                = "example"
  firewall_policy_arn = "arn:aws:network-firewall:us-east-1:123456789012:firewall-policy/example-policy"
  vpc_id              = "vpc-12345678"
  subnet_mapping {
    subnet_id = "subnet-12345678"
  }
}

resource "aws_networkfirewall_firewall" "networkfirewall_firewall_with_usage" {
  name                = "example"
  firewall_policy_arn = "arn:aws:network-firewall:us-east-1:123456789012:firewall-policy/example-policy"
  vpc_id              = "vpc-12345678"
  subnet_mapping {
    subnet_id = "subnet-12345678"
  }
}
