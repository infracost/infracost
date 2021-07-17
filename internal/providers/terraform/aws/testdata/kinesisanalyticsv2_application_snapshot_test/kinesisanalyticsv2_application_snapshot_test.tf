provider "aws" {
  region                      = "eu-west-2"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_get_ec2_platforms      = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_kinesisanalyticsv2_application" "flink" {
  name                   = "example-flink-application"
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.example.arn
}

resource "aws_kinesisanalyticsv2_application" "withoutUsage" {
  name                   = "example-flink-application"
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.example.arn
}

resource "aws_kinesisanalyticsv2_application_snapshot" "flink" {
  application_name = aws_kinesisanalyticsv2_application.flink.name
  snapshot_name    = "example-snapshot"
}

resource "aws_kinesisanalyticsv2_application_snapshot" "withoutUsage" {
  application_name = aws_kinesisanalyticsv2_application.withoutUsage.name
  snapshot_name    = "example-snapshot"
}


resource "aws_iam_role" "example" {
  name = "firehose_test_role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}