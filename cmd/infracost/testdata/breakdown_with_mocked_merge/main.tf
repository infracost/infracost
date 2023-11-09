provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

locals {
  value1 = data.terraform_remote_state.env

  value2 = merge(
    data.terraform_remote_state.env["test"],
    {
      instance_type = "m5.4xlarge"
    }
  )
}

resource "aws_instance" "web_app" {
  ami           = "ami-674cbc1e"
  instance_type = local.value2["instance_type"]
}
