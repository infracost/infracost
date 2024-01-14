provider "aws" {
  region                      = "eu-west-2"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_neptune_cluster" "default" {
  cluster_identifier                  = "neptune-cluster-demo"
  engine                              = "neptune"
  preferred_backup_window             = "07:00-09:00"
  skip_final_snapshot                 = true
  iam_database_authentication_enabled = true
  apply_immediately                   = true
}

resource "aws_neptune_cluster_instance" "dbR4LargeCount2" {
  count              = 2
  cluster_identifier = aws_neptune_cluster.default.id
  engine             = "neptune"
  instance_class     = "db.r4.large"
  apply_immediately  = true
}

resource "aws_neptune_cluster_instance" "dbT3Medium" {
  cluster_identifier = aws_neptune_cluster.default.id
  engine             = "neptune"
  instance_class     = "db.t3.medium"
  apply_immediately  = true
}

resource "aws_neptune_cluster_instance" "dbT3WithoutUsage" {
  cluster_identifier = aws_neptune_cluster.default.id
  engine             = "neptune"
  instance_class     = "db.t3.medium"
  apply_immediately  = true
}

resource "aws_neptune_cluster" "iooptimized" {
  cluster_identifier                  = "neptune-cluster-iooptimized"
  engine                              = "neptune"
  preferred_backup_window             = "07:00-09:00"
  storage_type                        = "iopt1"
  skip_final_snapshot                 = true
  iam_database_authentication_enabled = true
  apply_immediately                   = true
}

resource "aws_neptune_cluster_instance" "dbT3Medium-iooptimized" {
  cluster_identifier = aws_neptune_cluster.iooptimized.id
  engine             = "neptune"
  instance_class     = "db.t3.medium"
  apply_immediately  = true
}
