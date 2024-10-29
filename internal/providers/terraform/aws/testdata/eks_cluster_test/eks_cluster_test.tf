provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_eks_cluster" "my_aws_eks_cluster" {
  name     = "my_aws_eks_cluster"
  role_arn = "arn:aws:iam::123456789012:role/role"

  vpc_config {
    subnet_ids = ["subnet-123456"]
  }

  # Ensure that IAM Role permissions are created before and deleted after EKS Cluster handling.
  # Otherwise, EKS will not be able to properly delete EKS managed EC2 infrastructure such as Security Groups.
  depends_on = [
  ]
}


locals {
  versions = [
    "1.29",
    "1.28",
    "1.27",
    "1.26",
    "1.25",
    "1.24",
    "1.23",
  ]
}

resource "aws_eks_cluster" "extended_support" {
  for_each = { for version in local.versions : version => version }
  name     = "cluster"
  role_arn = "arn:aws:iam::123456789012:role/role"
  version  = each.value

  vpc_config {
    subnet_ids = ["subnet-123456"]
  }

  # Ensure that IAM Role permissions are created before and deleted after EKS Cluster handling.
  # Otherwise, EKS will not be able to properly delete EKS managed EC2 infrastructure such as Security Groups.
  depends_on = [
  ]
}
