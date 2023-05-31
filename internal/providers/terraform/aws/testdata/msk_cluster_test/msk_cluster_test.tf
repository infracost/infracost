provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
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
    instance_type   = "kafka.t3.small"
    security_groups = []
    storage_info {
      ebs_storage_info {
        volume_size = 500
      }
    }
  }
}

resource "aws_msk_cluster" "cluster-4-nodes" {
  cluster_name           = "cluster-4-nodes"
  kafka_version          = "2.4.1"
  number_of_broker_nodes = 4
  broker_node_group_info {
    client_subnets  = []
    instance_type   = "kafka.m5.24xlarge"
    security_groups = []
    storage_info {
      ebs_storage_info {
        volume_size = 1000
      }
    }
  }
}

resource "aws_appautoscaling_target" "autoscale_msk_cluster_target" {
  max_capacity       = 2000
  min_capacity       = 123
  resource_id        = aws_msk_cluster.cluster-autoscaling.arn
  scalable_dimension = "kafka:broker-storage:VolumeSize"
  service_namespace  = "kafka"
}

resource "aws_msk_cluster" "cluster-autoscaling" {
  cluster_name           = "cluster-autoscaling"
  kafka_version          = "2.4.1"
  number_of_broker_nodes = 2
  broker_node_group_info {
    client_subnets  = []
    instance_type   = "kafka.t3.small"
    security_groups = []
    storage_info {
      ebs_storage_info {
        volume_size = 1000
      }
    }
  }
}

resource "aws_appautoscaling_target" "autoscale_msk_cluster_target_usage" {
  max_capacity       = 2222
  min_capacity       = 1000
  resource_id        = aws_msk_cluster.cluster-autoscaling-usage.arn
  scalable_dimension = "kafka:broker-storage:VolumeSize"
  service_namespace  = "kafka"
}

resource "aws_msk_cluster" "cluster-autoscaling-usage" {
  cluster_name           = "cluster-autoscaling-usage"
  kafka_version          = "2.4.1"
  number_of_broker_nodes = 2
  broker_node_group_info {
    client_subnets  = []
    instance_type   = "kafka.t3.small"
    security_groups = []
    storage_info {
      ebs_storage_info {
        volume_size = 1000
      }
    }
  }
}
