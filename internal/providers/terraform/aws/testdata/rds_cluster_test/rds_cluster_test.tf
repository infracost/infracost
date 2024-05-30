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
  region                      = "sa-east-1"
  alias                       = "sa-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}


resource "aws_rds_cluster" "postgres_serverless" {
  cluster_identifier = "aurora-serverless"
  engine             = "aurora-postgresql"
  engine_mode        = "serverless"
  master_username    = "foo"
  master_password    = "barbut8chars"
}

resource "aws_rds_cluster" "my_sql_serverless" {
  cluster_identifier = "aurora-serverless"
  engine             = "aurora-mysql"
  engine_mode        = "serverless"
  master_username    = "foo"
  master_password    = "barbut8chars"
}

resource "aws_rds_cluster" "postgres_serverless_io_optimized" {
  cluster_identifier = "aurora-serverless"
  engine             = "aurora-postgresql"
  engine_mode        = "serverless"
  storage_type       = "aurora-iopt1"
  master_username    = "foo"
  master_password    = "barbut8chars"
}

resource "aws_rds_cluster" "my_sql_serverless_io_optimized" {
  cluster_identifier = "aurora-serverless"
  engine             = "aurora-mysql"
  engine_mode        = "serverless"
  storage_type       = "aurora-iopt1"
  master_username    = "foo"
  master_password    = "barbut8chars"
}

resource "aws_rds_cluster" "postgres_serverlessWithBackup" {
  cluster_identifier      = "aurora-serverless"
  engine                  = "aurora-postgresql"
  engine_mode             = "serverless"
  backup_retention_period = 5
  master_username         = "foo"
  master_password         = "barbut8chars"
}

resource "aws_rds_cluster" "postgres_serverlessWithExport" {
  cluster_identifier      = "aurora-serverless"
  engine                  = "aurora-postgresql"
  engine_mode             = "serverless"
  backup_retention_period = 5
  master_username         = "foo"
  master_password         = "barbut8chars"
}

resource "aws_rds_cluster" "mysql_backtrack" {
  cluster_identifier      = "aurora-mysql"
  engine                  = "aurora-mysql"
  backup_retention_period = 5
  master_username         = "foo"
  master_password         = "barbut8chars"
}

resource "aws_rds_cluster" "mysql_backtrack_withoutU" {
  cluster_identifier      = "aurora-mysql"
  engine                  = "aurora-mysql"
  backup_retention_period = 5
  master_username         = "foo"
  master_password         = "barbut8chars"
}

resource "aws_rds_cluster" "postgres_serverless_withoutU" {
  cluster_identifier = "aurora-serverless"
  engine             = "aurora-postgresql"
  engine_mode        = "serverless"
  master_username    = "foo"
  master_password    = "barbut8chars"
}

resource "aws_rds_cluster" "my_sql_brazil" {
  provider                = aws.sa-east-1
  cluster_identifier      = "aurora-mysql"
  engine                  = "aurora-mysql"
  backup_retention_period = 5
  master_username         = "foo"
  master_password         = "barbut8chars"
}
