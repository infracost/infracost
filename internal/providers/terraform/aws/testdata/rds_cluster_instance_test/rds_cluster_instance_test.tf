provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_rds_cluster" "default" {
  cluster_identifier = "aurora-cluster-demo"
  availability_zones = ["us-east-1a", "us-east-1b", "us-east-1c"]
  engine             = "aurora-mysql"
  database_name      = "mydb"
  master_username    = "foo"
  master_password    = "barbut8chars"
}

resource "aws_rds_cluster_instance" "cluster_instance" {
  identifier         = "aurora-cluster-demo"
  cluster_identifier = aws_rds_cluster.default.id
  instance_class     = "db.r4.large"
  engine             = aws_rds_cluster.default.engine
  engine_version     = aws_rds_cluster.default.engine_version
}

resource "aws_rds_cluster" "default_t3" {
  cluster_identifier = "aurora-cluster-demo"
  availability_zones = ["us-east-1a", "us-east-1b", "us-east-1c"]
  engine             = "aurora-mysql"
  database_name      = "mydb"
  master_username    = "foo"
  master_password    = "barbut8chars"
}

resource "aws_rds_cluster_instance" "cluster_instance_t3" {
  identifier         = "aurora-cluster-demo"
  cluster_identifier = aws_rds_cluster.default.id
  instance_class     = "db.t3.medium"
  engine             = aws_rds_cluster.default.engine
  engine_version     = aws_rds_cluster.default.engine_version
}

resource "aws_rds_cluster_instance" "cluster_instance_performance_insights" {
  identifier                            = "aurora-cluster-demo"
  cluster_identifier                    = aws_rds_cluster.default.id
  instance_class                        = "db.r4.8xlarge"
  engine                                = aws_rds_cluster.default.engine
  engine_version                        = aws_rds_cluster.default.engine_version
  performance_insights_enabled          = true
  performance_insights_retention_period = 731
}

resource "aws_rds_cluster_instance" "cluster_instance_performance_insights_with_usage" {
  identifier                            = "aurora-cluster-demo"
  cluster_identifier                    = aws_rds_cluster.default.id
  instance_class                        = "db.t3.large"
  engine                                = aws_rds_cluster.default.engine
  engine_version                        = aws_rds_cluster.default.engine_version
  performance_insights_enabled          = true
  performance_insights_retention_period = 731
}

resource "aws_rds_cluster_instance" "cluster_instance_1yr_no_upfront" {
  identifier         = "aurora-cluster-demo"
  cluster_identifier = aws_rds_cluster.default.id
  instance_class     = "db.r4.large"
  engine             = aws_rds_cluster.default.engine
  engine_version     = aws_rds_cluster.default.engine_version
}

resource "aws_rds_cluster_instance" "cluster_instance_1yr_partial_upfront" {
  identifier         = "aurora-cluster-demo"
  cluster_identifier = aws_rds_cluster.default.id
  instance_class     = "db.r4.large"
  engine             = aws_rds_cluster.default.engine
  engine_version     = aws_rds_cluster.default.engine_version
}

resource "aws_rds_cluster_instance" "cluster_instance_1yr_all_upfront" {
  identifier         = "aurora-cluster-demo"
  cluster_identifier = aws_rds_cluster.default.id
  instance_class     = "db.r4.large"
  engine             = aws_rds_cluster.default.engine
  engine_version     = aws_rds_cluster.default.engine_version
}

resource "aws_rds_cluster_instance" "cluster_instance_3yr_partial_upfront" {
  identifier         = "aurora-cluster-demo"
  cluster_identifier = aws_rds_cluster.default.id
  instance_class     = "db.r4.large"
  engine             = aws_rds_cluster.default.engine
  engine_version     = aws_rds_cluster.default.engine_version
}

resource "aws_rds_cluster_instance" "cluster_instance_3yr_all_upfront" {
  identifier         = "aurora-cluster-demo"
  cluster_identifier = aws_rds_cluster.default.id
  instance_class     = "db.t3.medium"
  engine             = aws_rds_cluster.default.engine
  engine_version     = aws_rds_cluster.default.engine_version
}

locals {
  extended_support_engined = {
    aurora-mysql = [
      "5.7",
      "5.7.44",
      "8.0",
      "8.0.36",
    ]
    aurora-postgresql = [
      "11",
      "11.22",
      "12",
      "13",
      "14",
      "15",
      "16"
    ]
  }
}

resource "aws_rds_cluster_instance" "extended_support" {
  for_each = { for entry in flatten([
    for engine, versions in local.extended_support_engined : [
      for version in versions : {
        engine  = engine
        version = version
      }
    ]
  ]) : "${entry.engine}-${entry.version}" => entry }

  identifier         = "aurora-cluster-demo"
  cluster_identifier = aws_rds_cluster.default.id

  engine         = each.value.engine
  engine_version = each.value.version
  instance_class = "db.t3.large"
}

resource "aws_rds_cluster" "serverless_v2" {
  cluster_identifier = "serverless-v2"
  engine             = "aurora-postgresql"
  engine_mode        = "provisioned"
  engine_version     = "13.6"
  database_name      = "test"
  master_username    = "test"
  master_password    = "must_be_eight_characters"
  storage_encrypted  = true

  serverlessv2_scaling_configuration {
    max_capacity = 1.0
    min_capacity = 0.5
  }
}

resource "aws_rds_cluster_instance" "serverless_v2_instance" {
  cluster_identifier = aws_rds_cluster.serverless_v2.id
  instance_class     = "db.serverless"
  engine             = aws_rds_cluster.serverless_v2.engine
  engine_version     = aws_rds_cluster.serverless_v2.engine_version
}
