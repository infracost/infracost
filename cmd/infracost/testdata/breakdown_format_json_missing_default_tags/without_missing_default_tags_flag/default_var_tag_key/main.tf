variable "tags" {
  type = map(string)
  default = {}
}

provider "aws" {
  region = "us-east-1"
  default_tags {
    tags = merge({
        Name = "web_app"
    }, var.tags)
  }
}

resource "aws_instance" "web_app" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge"
}
