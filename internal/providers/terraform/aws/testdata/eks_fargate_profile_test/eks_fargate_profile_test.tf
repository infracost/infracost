provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_eks_cluster" "example" {
  name     = "example"
  role_arn = "arn:aws:iam::123456789012:role/Example"

  vpc_config {
    subnet_ids = ["subnet_id"]
  }
}

resource "aws_eks_fargate_profile" "example" {
  cluster_name           = aws_eks_cluster.example.name
  fargate_profile_name   = "example"
  pod_execution_role_arn = "arn:aws:iam::123456789012:role/Example"
  subnet_ids             = ["subnet_id"]

  selector {
    namespace = "example"
  }
}
