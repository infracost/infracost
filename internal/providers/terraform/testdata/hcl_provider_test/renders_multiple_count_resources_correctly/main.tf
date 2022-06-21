provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_eip" "test" {
  count = 2
}

resource "aws_eip" "constant_string" {
  count = "1"
}

module "autos" {
  source = "./modules/autoscaling"

  types  = ["t2.micro", "t2.medium", "t2.large"]
  amount = 3
}
