provider "aws" {
  region = "us-west-2"
}

variable "instance_type" {}

resource "aws_instance" "web_app" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = var.instance_type
}

