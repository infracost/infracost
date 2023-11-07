provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

provider "aws" {
  alias                       = "ue2"
  region                      = "us-east-2"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}


resource "aws_docdb_cluster_instance" "db" {
  cluster_identifier = "fake123"
  instance_class     = "db.t3.medium"
}

resource "aws_docdb_cluster_instance" "medium" {
  cluster_identifier = "fake123"
  instance_class     = "db.t3.medium"
}

resource "aws_docdb_cluster_instance" "large" {
  cluster_identifier = "fake123"
  instance_class     = "db.r5.4xlarge"
}

resource "aws_docdb_cluster_instance" "dbue2" {
  provider           = aws.ue2
  cluster_identifier = "fake123"
  instance_class     = "db.t3.medium"
}

resource "aws_docdb_cluster_instance" "mediumue2" {
  provider           = aws.ue2
  cluster_identifier = "fake123"
  instance_class     = "db.t3.medium"
}

resource "aws_docdb_cluster_instance" "largeue2" {
  provider           = aws.ue2
  cluster_identifier = "fake123"
  instance_class     = "db.r5.4xlarge"
}

