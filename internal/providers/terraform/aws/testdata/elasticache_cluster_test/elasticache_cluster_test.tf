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

resource "aws_elasticache_cluster" "memcached" {
  cluster_id           = "cluster-example"
  engine               = "memcached"
  node_type            = "cache.m4.large"
  num_cache_nodes      = 2
  parameter_group_name = "default.redis3.2"
}

resource "aws_elasticache_cluster" "redis" {
  cluster_id           = "cluster-example"
  engine               = "redis"
  node_type            = "cache.m6g.12xlarge"
  num_cache_nodes      = 1
  parameter_group_name = "default.redis3.2"
}

resource "aws_elasticache_cluster" "redis_snapshot" {
  cluster_id               = "cluster-example"
  engine                   = "redis"
  node_type                = "cache.m6g.12xlarge"
  num_cache_nodes          = 1
  parameter_group_name     = "default.redis3.2"
  snapshot_retention_limit = 2
}

resource "aws_elasticache_cluster" "redis_snapshot_usage" {
  cluster_id               = "cluster-example"
  engine                   = "redis"
  node_type                = "cache.m6g.12xlarge"
  num_cache_nodes          = 1
  parameter_group_name     = "default.redis3.2"
  snapshot_retention_limit = 2
}