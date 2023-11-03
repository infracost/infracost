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

resource "aws_docdb_cluster_snapshot" "my_aws_docdb_cluster_snapshot" {
  db_cluster_identifier          = "fake"
  db_cluster_snapshot_identifier = "resourcetestsnapshot1234"
}

resource "aws_docdb_cluster_snapshot" "my_aws_docdb_cluster_snapshot_usage" {
  db_cluster_identifier          = "fake"
  db_cluster_snapshot_identifier = "resourcetestsnapshot1234"
}

resource "aws_docdb_cluster_snapshot" "my_aws_docdb_cluster_snapshot_ue2" {
  provider                       = aws.ue2
  db_cluster_identifier          = "fake"
  db_cluster_snapshot_identifier = "resourcetestsnapshot1234"
}

resource "aws_docdb_cluster_snapshot" "my_aws_docdb_cluster_snapshot_usage_ue2" {
  provider                       = aws.ue2
  db_cluster_identifier          = "fake"
  db_cluster_snapshot_identifier = "resourcetestsnapshot1234"
}
