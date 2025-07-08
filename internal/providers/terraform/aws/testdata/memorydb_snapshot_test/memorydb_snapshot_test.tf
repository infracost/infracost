provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_memorydb_snapshot" "example" {
  cluster_name = aws_memorydb_cluster.example.name
  name         = "example-snapshot"
}

resource "aws_memorydb_cluster" "example" {
  name               = "example-cluster"
  node_type          = "db.r6g.large"
  num_shards         = 1
  num_replicas_per_shard = 1
}
