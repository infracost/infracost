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

resource "aws_msk_cluster" "cluster-2-nodes" {
  cluster_name           = "cluster-2-nodes"
  kafka_version          = "2.4.1"
  number_of_broker_nodes = 2
  broker_node_group_info {
    client_subnets  = []
    ebs_volume_size = 500
    instance_type   = "kafka.t3.small"
    security_groups = []
  }
}

resource "aws_msk_cluster" "cluster-4-nodes" {
  cluster_name           = "cluster-4-nodes"
  kafka_version          = "2.4.1"
  number_of_broker_nodes = 4
  broker_node_group_info {
    client_subnets  = []
    ebs_volume_size = 1000
    instance_type   = "kafka.m5.24xlarge"
    security_groups = []
  }
}