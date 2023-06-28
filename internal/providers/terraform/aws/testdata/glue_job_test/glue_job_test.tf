provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_glue_job" "job_no_usage" {
  name     = "example job no usage"
  role_arn = "arn:aws:glue:us-east-1:123456789012:resource-id"

  command {
    script_location = "s3://bucket.test/example.script"
  }
}

resource "aws_glue_job" "python_shell" {
  name         = "example job no usage"
  role_arn     = "arn:aws:glue:us-east-1:123456789012:resource-id"
  max_capacity = "0.0625"

  command {
    name            = "pythonshell"
    script_location = "s3://bucket.test/example.script"
  }
}

resource "aws_glue_job" "cap_10" {
  name         = "example streaming job"
  role_arn     = "arn:aws:glue:us-east-1:123456789012:resource-id"
  max_capacity = 10

  command {
    script_location = "s3://bucket.test/example.script"
  }
}

resource "aws_glue_job" "workers_default" {
  name              = "example streaming job"
  role_arn          = "arn:aws:glue:us-east-1:123456789012:resource-id"
  number_of_workers = 10

  command {
    script_location = "s3://bucket.test/example.script"
  }
}

resource "aws_glue_job" "workers_g1x" {
  name              = "example streaming job"
  role_arn          = "arn:aws:glue:us-east-1:123456789012:resource-id"
  number_of_workers = 10
  worker_type       = "G.1X"

  command {
    script_location = "s3://bucket.test/example.script"
  }
}

resource "aws_glue_job" "workers_g2x" {
  name              = "example streaming job"
  role_arn          = "arn:aws:glue:us-east-1:123456789012:resource-id"
  number_of_workers = 10
  worker_type       = "G.2X"

  command {
    script_location = "s3://bucket.test/example.script"
  }
}
