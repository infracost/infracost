provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
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


resource "aws_elasticache_cluster" "redis_reserved_1yr_no_upfront" {
  cluster_id           = "cluster-example"
  engine               = "redis"
  node_type            = "cache.m6g.12xlarge"
  num_cache_nodes      = 1
  parameter_group_name = "default.redis3.2"
}

resource "aws_elasticache_cluster" "redis_reserved_1yr_partial_upfront" {
  cluster_id           = "cluster-example"
  engine               = "redis"
  node_type            = "cache.m6g.12xlarge"
  num_cache_nodes      = 1
  parameter_group_name = "default.redis3.2"
}

resource "aws_elasticache_cluster" "redis_reserved_1yr_all_upfront" {
  cluster_id           = "cluster-example"
  engine               = "redis"
  node_type            = "cache.m6g.12xlarge"
  num_cache_nodes      = 1
  parameter_group_name = "default.redis3.2"
}

resource "aws_elasticache_replication_group" "replication_group" {
  description                = "This Replication Group"
  replication_group_id       = "tf-rep-group-1"
  automatic_failover_enabled = true
  node_type                  = "cache.m4.large"

  engine = "redis"

  num_node_groups         = 4
  replicas_per_node_group = 3
}

resource "aws_elasticache_cluster" "replication_group" {
  cluster_id           = "cluster-example"
  replication_group_id = aws_elasticache_replication_group.replication_group.id
}
