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

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "my_aws_subnet" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.1.0/24"

  tags = {
    Name = "Main"
  }
}

resource "aws_iam_role" "my_aws_iam_role" {
    name = "awsconfig-example"

    assume_role_policy = <<POLICY
{}
POLICY
}

resource "aws_eks_cluster" "my_aws_eks_cluster" {
  name     = "my_aws_eks_cluster"
  role_arn = aws_iam_role.my_aws_iam_role.arn

  vpc_config {
    subnet_ids = [aws_subnet.my_aws_subnet.id]
  }

  # Ensure that IAM Role permissions are created before and deleted after EKS Cluster handling.
  # Otherwise, EKS will not be able to properly delete EKS managed EC2 infrastructure such as Security Groups.
  depends_on = [
  ]
}
