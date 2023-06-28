provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_opensearch_domain" "gp2" {
  domain_name = "example-domain"

  cluster_config {
    instance_type            = "c4.2xlarge.search"
    instance_count           = 3
    dedicated_master_enabled = true
    dedicated_master_type    = "c4.8xlarge.search"
    dedicated_master_count   = 1
    warm_enabled             = true
    warm_count               = 2
    warm_type                = "ultrawarm1.medium.search"
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 400
    volume_type = "gp2"
  }
}

resource "aws_opensearch_domain" "io1" {
  domain_name = "example-domain"

  cluster_config {
    instance_type  = "c4.2xlarge.search"
    instance_count = 3
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 1000
    volume_type = "io1"
    iops        = 10
  }
}

resource "aws_opensearch_domain" "std" {
  domain_name = "example-domain"

  cluster_config {
    instance_type  = "c4.2xlarge.search"
    instance_count = 3
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 123
    volume_type = "standard"
  }
}
