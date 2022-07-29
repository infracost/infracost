provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

module "this" {
  count         = 2
  source        = "./modules/test"
  instance_type = count.index == 0 ? "m5.4xlarge" : "t2.micro"
}
