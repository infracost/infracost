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
resource "aws_iam_role" "my_aws_iam_role" {
  name               = "awsconfig-example"
  assume_role_policy = <<POLICY
{}
POLICY
}
resource "aws_lambda_function" "lambda" {
  function_name = "lambda_function_name"
  role          = "arn:aws:lambda:us-east-1:account-id:resource-id"
  handler       = "exports.test"
  runtime       = "nodejs12.x"
}
resource "aws_sfn_state_machine" "express" {
  name       = "my-state-machine"
  role_arn   = aws_iam_role.my_aws_iam_role.arn
  type       = "EXPRESS"
  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Task",
      "Resource": "${aws_lambda_function.lambda.arn}",
      "End": true
    }
  }
}
EOF
}

resource "aws_sfn_state_machine" "standard" {
  name       = "my-state-machine"
  role_arn   = aws_iam_role.my_aws_iam_role.arn
  type       = "STANDARD"
  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Task",
      "Resource": "${aws_lambda_function.lambda.arn}",
      "End": true
    }
  }
}
EOF
}