provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_elasticsearch_domain" "gp2" {
  domain_name           = "example-domain"
  elasticsearch_version = "1.5"

  cluster_config {
    instance_type            = "c4.2xlarge.elasticsearch"
    instance_count           = 3
    dedicated_master_enabled = true
    dedicated_master_type    = "c4.8xlarge.elasticsearch"
    dedicated_master_count   = 1
    warm_enabled             = true
    warm_count               = 2
    warm_type                = "ultrawarm1.medium.elasticsearch"
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 400
    volume_type = "gp2"
  }
}

resource "aws_elasticsearch_domain" "gp3" {
  domain_name           = "example-domain"
  elasticsearch_version = "1.5"

  cluster_config {
    instance_type            = "c4.2xlarge.elasticsearch"
    instance_count           = 3
    dedicated_master_enabled = true
    dedicated_master_type    = "c4.8xlarge.elasticsearch"
    dedicated_master_count   = 1
    warm_enabled             = true
    warm_count               = 2
    warm_type                = "ultrawarm1.medium.elasticsearch"
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 400
    volume_type = "gp3"
  }
}

resource "aws_elasticsearch_domain" "gp3_throughput" {
  domain_name           = "example-domain"
  elasticsearch_version = "1.5"

  for_each = {
    "below_min_storage_free" : {
      instance_count = 1,
      storage : 150,
      throughput : 125,
    }
    "below_min_storage_paid" : {
      instance_count = 1,
      storage : 150,
      throughput : 200,
    }
    "below_min_storage_mul_paid" : {
      instance_count = 2,
      storage : 150,
      throughput : 200,
    }
    "above_min_storage_free" : {
      instance_count = 1,
      storage : 180,
      throughput : 125,
    }
    "above_3TB_free" : {
      instance_count = 1,
      storage : 3500,
      throughput : 250,
    }
    "above_3TB_free_round_up" : {
      instance_count = 1,
      storage : 5000,
      throughput : 500,
    }
    "above_3TB_paid" : {
      instance_count = 1,
      storage : 4000,
      throughput : 700,
    }
    "above_3TB_paid_mul" : {
      instance_count = 2,
      storage : 4000,
      throughput : 700,
    }
  }

  cluster_config {
    instance_type  = "c4.2xlarge.elasticsearch"
    instance_count = each.value.instance_count
  }

  ebs_options {
    ebs_enabled = true
    throughput  = each.value.throughput
    volume_size = each.value.storage
    volume_type = "gp3"
  }
}

resource "aws_elasticsearch_domain" "io1" {
  domain_name           = "example-domain"
  elasticsearch_version = "1.5"

  cluster_config {
    instance_type  = "c4.2xlarge.elasticsearch"
    instance_count = 3
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 1000
    volume_type = "io1"
    iops        = 10
  }
}

resource "aws_elasticsearch_domain" "std" {
  domain_name           = "example-domain"
  elasticsearch_version = "1.5"

  cluster_config {
    instance_type  = "c4.2xlarge.elasticsearch"
    instance_count = 3
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 123
    volume_type = "standard"
  }
}
