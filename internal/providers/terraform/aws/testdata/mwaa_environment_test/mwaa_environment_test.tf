provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_mwaa_environment" "default_environment_class" {
  dag_s3_path        = "dags/"
  execution_role_arn = "arn:aws:iam::123456789012:role/role"
  name               = "example"

  network_configuration {
    security_group_ids = ["sg-903004f8"]
    subnet_ids         = ["subnet-123456", "subnet-654321"]
  }

  source_bucket_arn = "arn:aws:s3:::bucket"
}

resource "aws_mwaa_environment" "medium" {
  dag_s3_path        = "dags/"
  execution_role_arn = "arn:aws:iam::123456789012:role/role"
  name               = "example"
  environment_class  = "mw1.medIUM"

  network_configuration {
    security_group_ids = ["sg-903004f8"]
    subnet_ids         = ["subnet-123456", "subnet-654321"]
  }

  source_bucket_arn = "arn:aws:s3:::bucket"
}

resource "aws_mwaa_environment" "large_with_usage" {
  dag_s3_path        = "dags/"
  execution_role_arn = "arn:aws:iam::123456789012:role/role"
  name               = "example"
  environment_class  = "mW1.laRgE"

  network_configuration {
    security_group_ids = ["sg-903004f8"]
    subnet_ids         = ["subnet-123456", "subnet-654321"]
  }

  source_bucket_arn = "arn:aws:s3:::bucket"
}
