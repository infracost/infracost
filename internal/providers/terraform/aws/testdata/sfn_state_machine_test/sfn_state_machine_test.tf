provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_sfn_state_machine" "expressWithoutUsage" {
  name       = "my-state-machine"
  role_arn   = "arn:aws:lambda:us-east-1:123456789012:resource-id"
  type       = "EXPRESS"
  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Task",
      "Resource": "fake123",
      "End": true
    }
  }
}
EOF
}

resource "aws_sfn_state_machine" "express1Tier" {
  name       = "my-state-machine"
  role_arn   = "arn:aws:lambda:us-east-1:123456789012:resource-id"
  type       = "EXPRESS"
  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Task",
      "Resource": "fake123",
      "End": true
    }
  }
}
EOF
}

resource "aws_sfn_state_machine" "express2Tiers" {
  name       = "my-state-machine"
  role_arn   = "arn:aws:lambda:us-east-1:123456789012:resource-id"
  type       = "EXPRESS"
  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Task",
      "Resource": "fake123",
      "End": true
    }
  }
}
EOF
}

resource "aws_sfn_state_machine" "express3Tiers" {
  name       = "my-state-machine"
  role_arn   = "arn:aws:lambda:us-east-1:123456789012:resource-id"
  type       = "EXPRESS"
  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Task",
      "Resource": "fake123",
      "End": true
    }
  }
}
EOF
}

resource "aws_sfn_state_machine" "standardWithoutUsage" {
  name       = "my-state-machine"
  role_arn   = "arn:aws:lambda:us-east-1:123456789012:resource-id"
  type       = "STANDARD"
  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Task",
      "Resource": "fake123",
      "End": true
    }
  }
}
EOF
}

resource "aws_sfn_state_machine" "standard" {
  name       = "my-state-machine"
  role_arn   = "arn:aws:lambda:us-east-1:123456789012:resource-id"
  type       = "STANDARD"
  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Task",
      "Resource": "fake123",
      "End": true
    }
  }
}
EOF
}
