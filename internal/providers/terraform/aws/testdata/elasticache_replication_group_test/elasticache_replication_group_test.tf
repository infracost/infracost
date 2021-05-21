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

resource "aws_elasticache_replication_group" "cluster" {
  replication_group_description = "This Replication Group"
  replication_group_id          = "tf-rep-group-1"
  automatic_failover_enabled    = true
  node_type                     = "cache.m4.large"

  engine = "redis"

  cluster_mode {
    num_node_groups         = 4
    replicas_per_node_group = 3
  }
}

resource "aws_elasticache_replication_group" "non-cluster" {
  replication_group_description = "This Replication Group"
  replication_group_id          = "tf-rep-group-2"

  engine = "redis"

  node_type             = "cache.r5.4xlarge"
  number_cache_clusters = 3
}

resource "aws_elasticache_replication_group" "non-cluster-snapshot" {
  replication_group_description = "This Replication Group"
  replication_group_id          = "tf-rep-group-3"
  snapshot_retention_limit      = 2

  engine = "redis"

  node_type             = "cache.m6g.12xlarge"
  number_cache_clusters = 3
}