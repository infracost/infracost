provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

# Add example resources for Ec2Host below
resource "aws_ec2_host" "ec2_host" {
  instance_type     = "c5.18xlarge"
  availability_zone = "us-west-2a"
  host_recovery     = "on"
  auto_placement    = "on"
}

resource "aws_ec2_host" "host-1yr-no-upfront" {
  instance_family   = "m5"
  availability_zone = "us-east-2a"
  host_recovery     = "on"
  auto_placement    = "on"
}

resource "aws_ec2_host" "host-1yr-partial-upfront" {
  instance_family   = "c5"
  availability_zone = "us-east-2a"
  host_recovery     = "on"
  auto_placement    = "on"
}

resource "aws_ec2_host" "host-1yr-all-upfront" {
  instance_family   = "m5d"
  availability_zone = "us-east-2a"
  host_recovery     = "on"
  auto_placement    = "on"
}

resource "aws_ec2_host" "host-3yr-no-upfront" {
  instance_family   = "m5"
  availability_zone = "us-east-2a"
  host_recovery     = "on"
  auto_placement    = "on"
}

resource "aws_ec2_host" "host-3yr-partial-upfront" {
  instance_family   = "c5"
  availability_zone = "us-east-2a"
  host_recovery     = "on"
  auto_placement    = "on"
}

resource "aws_ec2_host" "host-3yr-all-upfront" {
  instance_family   = "m5d"
  availability_zone = "us-east-2a"
  host_recovery     = "on"
  auto_placement    = "on"
}

resource "aws_ec2_host" "mac" {
  instance_type     = "mac1.metal"
  availability_zone = "us-east-2a"
}
