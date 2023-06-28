provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_ebs_volume" "gp2" {
  availability_zone = "us-east-1a"
  size              = 10
}

resource "aws_ebs_volume" "standard" {
  availability_zone = "us-east-1a"
  size              = 20
  type              = "standard"
}

resource "aws_ebs_volume" "io1" {
  availability_zone = "us-east-1a"
  type              = "io1"
  size              = 30
  iops              = 300
}

resource "aws_ebs_volume" "io2" {
  availability_zone = "us-east-1a"
  type              = "io2"
  size              = 30
  iops              = 300
}

resource "aws_ebs_volume" "st1" {
  availability_zone = "us-east-1a"
  size              = 40
  type              = "st1"
}

resource "aws_ebs_volume" "sc1" {
  availability_zone = "us-east-1a"
  size              = 50
  type              = "sc1"
}

resource "aws_ebs_volume" "gp3" {
  availability_zone = "us-west-2a"
  size              = 40
  type              = "gp3"
  iops              = 4000
  throughput        = 130

  tags = {
    Name = "HelloWorld"
  }
}

resource "aws_ebs_volume" "standard_withUsage" {
  availability_zone = "us-east-1a"
  size              = 20
  type              = "standard"
}
