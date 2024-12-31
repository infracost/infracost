variable "tags" {
  type = map(string)
}

variable "aws_region" {
  type = string
}

provider "aws" {
  region = var.aws_region
  default_tags {
    tags = var.tags
  }
}

resource "aws_instance" "web_app" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge"
}
