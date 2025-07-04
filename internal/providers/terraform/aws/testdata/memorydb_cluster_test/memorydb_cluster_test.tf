provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_memorydb_cluster" "redis" {
  name               = "my-cluster"
  node_type          = "db.r6g.large"
  num_shards         = 2
  num_replicas_per_shard = 1
  engine             = "redis"
  snapshot_retention_limit = 2
}

resource "aws_memorydb_cluster" "valkey" {
  name               = "my-valkey-cluster"
  node_type          = "db.r6g.large"
  num_shards         = 2
  num_replicas_per_shard = 1
  engine             = "valkey"
  snapshot_retention_limit = 2
}

resource "aws_memorydb_cluster" "redis_with_usage" {
  name               = "my-cluster-with-usage"
  node_type          = "db.r6g.large"
  num_shards         = 2
  num_replicas_per_shard = 1
  engine             = "redis"
  snapshot_retention_limit = 2
}

resource "aws_memorydb_cluster" "valkey_with_usage" {
  name               = "my-valkey-cluster-with-usage"
  node_type          = "db.r6g.large"
  num_shards         = 2
  num_replicas_per_shard = 1
  engine             = "valkey"
  snapshot_retention_limit = 2
}

resource "aws_memorydb_cluster" "redis_reserved" {
  name               = "my-reserved-cluster"
  node_type          = "db.r6g.large"
  num_shards         = 2
  num_replicas_per_shard = 1
  engine             = "redis"
  snapshot_retention_limit = 2
}

resource "aws_appautoscaling_target" "memorydb_shards" {
  max_capacity       = 4
  min_capacity       = 2
  resource_id        = "memorydb:my-autoscaling-cluster"
  scalable_dimension = "memorydb:cluster:Shards"
  service_namespace  = "memorydb"
}

resource "aws_appautoscaling_target" "memorydb_replicas" {
  max_capacity       = 2
  min_capacity       = 1
  resource_id        = "memorydb:my-autoscaling-cluster"
  scalable_dimension = "memorydb:cluster:ReplicasPerShard"
  service_namespace  = "memorydb"
}

resource "aws_memorydb_cluster" "autoscaling" {
  name               = "my-autoscaling-cluster"
  node_type          = "db.r6g.large"
  num_shards         = 2
  num_replicas_per_shard = 1
  engine             = "redis"
  snapshot_retention_limit = 2
}
