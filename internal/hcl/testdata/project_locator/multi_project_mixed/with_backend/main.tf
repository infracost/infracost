terraform {
  backend "remote" {
    hostname     = "app.terraform.io"
    organization = "example_corp"

    workspaces {
      name = "example_corp/web-app-prod"
    }
  }
}

resource "aws_instance" "web_app" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge"

  root_block_device {
    volume_size = 50
  }
}

