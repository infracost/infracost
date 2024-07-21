provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

variable "instance_config" {
  type = map(any)
  default = {
    instance_type = "m5.4xlarge"
    storage       = 10
  }
}

variable "lambda_configs" {
  type = list(object({
    memory_size = number
  }))
  default = [
    {
      memory_size = 1024
    }
  ]
}


resource "aws_instance" "web_app" {
  ami           = "ami-674cbc1e"
  instance_type = var.instance_config.instance_type

  root_block_device {
    volume_size = var.instance_config.storage
  }
}

resource "aws_lambda_function" "hello_world" {
  count         = length(var.lambda_configs)
  function_name = "hello_world"
  role          = "arn:aws:lambda:us-east-1:aws:resource-id"
  handler       = "exports.test"
  runtime       = "nodejs12.x"
  filename      = "function.zip"
  memory_size   = var.lambda_configs[count.index].memory_size
}
