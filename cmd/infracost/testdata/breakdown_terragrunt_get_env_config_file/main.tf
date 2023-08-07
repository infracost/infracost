variable "env" {
  default = "dev"
}

provider "aws" {
  region                      = var.region
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_instance" "web_app" {
  ami           = "ami-674cbc1e"
  instance_type = var.env == "prod" ? "t2.medium" : "t2.micro"

  root_block_device {
    volume_size = 50
  }
}
