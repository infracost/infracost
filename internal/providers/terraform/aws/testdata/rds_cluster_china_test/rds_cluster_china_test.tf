provider "aws" {
  region                      = "cn-north-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_rds_cluster" "mysql_china" {
  cluster_identifier      = "aurora-mysql"
  engine                  = "aurora-mysql"
  backup_retention_period = 5
  master_username         = "foo"
  master_password         = "barbut8chars"
}
