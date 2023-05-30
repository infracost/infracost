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

resource "aws_ebs_snapshot" "gp2" {
  volume_id = aws_ebs_volume.gp2.id
}

resource "aws_ebs_snapshot_copy" "gp2" {
  source_snapshot_id = aws_ebs_snapshot.gp2.id
  source_region      = "us-east-1"
}
