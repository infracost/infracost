provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

locals {
  value2 = merge(
    data.terraform_remote_state.env["test"],
    {
      instance_type = "m5.4xlarge"
    }
  )

  value3 = merge(
    data.terraform_remote_state.env.test,
    {
      tags = merge(
        data.terraform_remote_state.env.test["foo"]["bar"],
        {
          instance_type = "m5.4xlarge"
        }
      )
    }
  )
}

resource "aws_instance" "web_app" {
  ami           = "ami-674cbc1e"
  instance_type = local.value2["instance_type"]
}
#
resource "aws_instance" "web_app2" {
  ami           = "ami-674cbc1e"
  instance_type = local.value3["tags"]["instance_type"]
}
