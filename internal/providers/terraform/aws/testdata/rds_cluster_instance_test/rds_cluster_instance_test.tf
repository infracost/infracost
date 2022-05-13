provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_get_ec2_platforms      = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_rds_cluster" "default" {
  cluster_identifier = "aurora-cluster-demo"
  availability_zones = ["us-east-1a", "us-east-1b", "us-east-1c"]
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
