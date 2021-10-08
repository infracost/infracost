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

resource "aws_iam_role" "r" {
  name = "awsconfig-example"

  assume_role_policy = <<POLICY
{}
POLICY
}

resource "aws_security_group" "example" {
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "private" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.1.0/24"

  tags = {
    Name = "Main"
  }
}

resource "aws_s3_bucket" "example" {
  bucket = "mybucket"
  acl    = "private"

  tags = {
    Name = "My bucket"
  }
}

resource "aws_mwaa_environment" "default_environment_class" {
  dag_s3_path        = "dags/"
  execution_role_arn = aws_iam_role.r.arn
  name               = "example"

  network_configuration {
    security_group_ids = [aws_security_group.example.id]
    subnet_ids         = aws_subnet.private[*].id
  }

  source_bucket_arn = aws_s3_bucket.example.arn
}

resource "aws_mwaa_environment" "medium" {
  dag_s3_path        = "dags/"
  execution_role_arn = aws_iam_role.r.arn
  name               = "example"
  environment_class  = "mw1.medIUM"

  network_configuration {
    security_group_ids = [aws_security_group.example.id]
    subnet_ids         = aws_subnet.private[*].id
  }

  source_bucket_arn = aws_s3_bucket.example.arn
}

resource "aws_mwaa_environment" "large_with_usage" {
  dag_s3_path        = "dags/"
  execution_role_arn = aws_iam_role.r.arn
  name               = "example"
  environment_class  = "mW1.laRgE"

  network_configuration {
    security_group_ids = [aws_security_group.example.id]
    subnet_ids         = aws_subnet.private[*].id
  }

  source_bucket_arn = aws_s3_bucket.example.arn
}