variable "range" {}

resource "aws_lambda_function" "test" {
  for_each = var.range

  function_name = "hello_world"
  role          = "arn:aws:lambda:us-east-1:aws:resource-id"
  handler       = "exports.test"
  runtime       = "nodejs12.x"
  filename      = "function.zip"
  memory_size   = 1024
}

resource "aws_lambda_function" "test2" {
  function_name = "hello_world"
  role          = "arn:aws:lambda:us-east-1:aws:resource-id"
  handler       = "exports.test"
  runtime       = "nodejs12.x"
  filename      = "function.zip"
  memory_size   = 1024
}
