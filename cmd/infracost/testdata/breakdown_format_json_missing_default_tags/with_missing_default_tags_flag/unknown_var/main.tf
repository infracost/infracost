provider "aws" {
  region = "us-east-1"
  default_tags {
    tags = var.unknown_var
  }
}

resource "aws_instance" "web_app" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge"
}
