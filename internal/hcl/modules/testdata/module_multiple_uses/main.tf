provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

module "registry-module-1" {
  source  = "terraform-aws-modules/ec2-instance/aws"
  version = "3.4.0"

  name = "my-instance"

  ami                    = "ami-ebd02392"
  instance_type          = "t2.micro"
  key_name               = "user1"
  monitoring             = true
  vpc_security_group_ids = ["sg-12345678"]
  subnet_id              = "subnet-eddcdzz4"
}

module "registry-module-2" {
  source  = "terraform-aws-modules/ec2-instance/aws"
  version = "3.4.0"

  name = "my-instance"

  ami                    = "ami-ebd02392"
  instance_type          = "t2.small"
  key_name               = "user1"
  monitoring             = true
  vpc_security_group_ids = ["sg-12345678"]
  subnet_id              = "subnet-eddcdzz4"
}

