provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_elasticache_replication_group" "cluster" {
  description                = "This Replication Group"
  replication_group_id       = "tf-rep-group-1"
  automatic_failover_enabled = true
  node_type                  = "cache.m4.large"

  engine = "redis"

  num_node_groups         = 4
  replicas_per_node_group = 3
}

resource "aws_elasticache_replication_group" "non-cluster" {
  description          = "This Replication Group"
  replication_group_id = "tf-rep-group-2"

  engine = "redis"

  node_type          = "cache.r5.4xlarge"
  num_cache_clusters = 3
}

resource "aws_elasticache_replication_group" "non-cluster-snapshot" {
  description              = "This Replication Group"
  replication_group_id     = "tf-rep-group-3"
  snapshot_retention_limit = 2

  engine = "redis"

  node_type          = "cache.m6g.12xlarge"
  num_cache_clusters = 3
}

resource "aws_elasticache_replication_group" "cluster-autoscale" {
  description                = "This Replication Group"
  replication_group_id       = "tf-rep-group-4"
  automatic_failover_enabled = true
  node_type                  = "cache.m4.large"

  engine = "redis"

  num_node_groups         = 4
  replicas_per_node_group = 3
}

resource "aws_appautoscaling_target" "autoscale_node_groups" {
  max_capacity       = 13
  min_capacity       = 8
  resource_id        = "replication-group/${aws_elasticache_replication_group.cluster-autoscale.replication_group_id}"
  scalable_dimension = "elasticache:replication-group:NodeGroups"
  service_namespace  = "elasticache"
}

resource "aws_appautoscaling_target" "autoscale_replicas" {
  max_capacity       = 17
  min_capacity       = 5
  resource_id        = "replication-group/${aws_elasticache_replication_group.cluster-autoscale.replication_group_id}"
  scalable_dimension = "elasticache:replication-group:Replicas"
  service_namespace  = "elasticache"
}

resource "aws_elasticache_replication_group" "cluster-autoscale-usage" {
  description                = "This Replication Group"
  replication_group_id       = "tf-rep-group-7"
  automatic_failover_enabled = true
  node_type                  = "cache.m4.large"

  engine = "redis"

  num_node_groups         = 4
  replicas_per_node_group = 3
}

resource "aws_appautoscaling_target" "autoscale_node_groups_usage" {
  max_capacity       = 2
  min_capacity       = 1
  resource_id        = "replication-group/tf-rep-group-7"
  scalable_dimension = "elasticache:replication-group:NodeGroups"
  service_namespace  = "elasticache"
}

resource "aws_elasticache_replication_group" "cluster_reserved" {
  description          = "This Replication Group"
  replication_group_id = "tf-rep-group-2"

  engine = "redis"

  node_type          = "cache.m6g.12xlarge"
  num_cache_clusters = 3
}
