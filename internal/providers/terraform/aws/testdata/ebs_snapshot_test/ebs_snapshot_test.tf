provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_ebs_volume" "gp2" {
  availability_zone = "us-east-1a"
  size              = 10
}

resource "aws_ebs_snapshot" "gp2" {
  volume_id = aws_ebs_volume.gp2.id
}

resource "aws_ebs_snapshot" "gp2_usage" {
  volume_id = "fake"
}