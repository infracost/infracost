provider "aws" {
  region                      = "eu-west-2"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}


resource "aws_neptune_cluster" "fiveDaysRetentionPeriod" {
  cluster_identifier                  = "neptune-cluster-demo"
  engine                              = "neptune"
  backup_retention_period             = 5
  preferred_backup_window             = "07:00-09:00"
  skip_final_snapshot                 = true
  iam_database_authentication_enabled = true
  apply_immediately                   = true
}

resource "aws_neptune_cluster_snapshot" "fiveDaysRetenPeriod" {
  db_cluster_identifier          = aws_neptune_cluster.fiveDaysRetentionPeriod.id
  db_cluster_snapshot_identifier = "resourcetestsnapshot1234"
}


resource "aws_neptune_cluster" "oneDayRetentionPeriod" {
  cluster_identifier                  = "neptune-cluster-demo"
  engine                              = "neptune"
  preferred_backup_window             = "07:00-09:00"
  skip_final_snapshot                 = true
  iam_database_authentication_enabled = true
  apply_immediately                   = true
}

resource "aws_neptune_cluster_snapshot" "oneDayRetentionPeriod" {
  db_cluster_identifier          = aws_neptune_cluster.oneDayRetentionPeriod.id
  db_cluster_snapshot_identifier = "resourcetestsnapshot1234"
}


resource "aws_neptune_cluster" "withoutUsage" {
  cluster_identifier                  = "neptune-cluster-demo"
  engine                              = "neptune"
  preferred_backup_window             = "07:00-09:00"
  backup_retention_period             = 5
  skip_final_snapshot                 = true
  iam_database_authentication_enabled = true
  apply_immediately                   = true
}

resource "aws_neptune_cluster_snapshot" "withoutUsage" {
  db_cluster_identifier          = aws_neptune_cluster.withoutUsage.id
  db_cluster_snapshot_identifier = "resourcetestsnapshot1234"
}

resource "aws_neptune_cluster_snapshot" "inAnotherModule" {
  db_cluster_identifier          = "in-another-module"
  db_cluster_snapshot_identifier = "resourcetestsnapshot1234"
}
