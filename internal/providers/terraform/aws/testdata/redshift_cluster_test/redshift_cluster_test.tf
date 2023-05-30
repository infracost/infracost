provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_redshift_cluster" "ca" {
  cluster_identifier = "tf-ca-cluster"
  database_name      = "mydb"
  master_username    = "foo"
  master_password    = "Mustbe8characters"
  node_type          = "dc2.large"
  cluster_type       = "multi-node"
  number_of_nodes    = "4"
}

resource "aws_redshift_cluster" "ca_withUsage" {
  cluster_identifier = "tf-ca-withusage-cluster"
  database_name      = "mydb"
  master_username    = "foo"
  master_password    = "Mustbe8characters"
  node_type          = "ds2.8xlarge"
  cluster_type       = "single-node"
}

resource "aws_redshift_cluster" "manageda" {
  cluster_identifier = "tf-manageda-cluster"
  database_name      = "mydb"
  master_username    = "foo"
  master_password    = "Mustbe8characters"
  node_type          = "ra3.4xlarge"
  cluster_type       = "multi-node"
  number_of_nodes    = "6"
}

resource "aws_redshift_cluster" "manageda_withUsage" {
  cluster_identifier = "tf-manageda-withusage-cluster"
  database_name      = "mydb"
  master_username    = "foo"
  master_password    = "Mustbe8characters"
  node_type          = "ra3.16xlarge"
  cluster_type       = "single-node"
}
