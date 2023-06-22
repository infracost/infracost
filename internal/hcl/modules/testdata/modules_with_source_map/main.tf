provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

module "registry-module" {
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

module "git-module-with-ref" {
  source  = "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance.git?ref=v4.5.0"
  version = "4.5.0"

  name = "my-instance"

  ami                    = "ami-ebd02392"
  instance_type          = "t2.micro"
  key_name               = "user1"
  monitoring             = true
  vpc_security_group_ids = ["sg-12345678"]
  subnet_id              = "subnet-eddcdzz4"
}

module "git-module-with-mapped-ref" {
  source = "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance.git?ref=v4.4.0"

  name = "my-instance"

  ami                    = "ami-ebd02392"
  instance_type          = "t2.micro"
  key_name               = "user1"
  monitoring             = true
  vpc_security_group_ids = ["sg-12345678"]
  subnet_id              = "subnet-eddcdzz4"
}


module "git-module-not-replaced" {
  source = "git::https://github.com/terraform-aws-modules/terraform-aws-sns.git"

  name = "my-topic"
}

module "registry-module-with-submodule" {
  source  = "terraform-aws-modules/route53/aws//modules/zones"
  version = "2.5.0"

  zones = {
    "example.com" = {
      comment = "example.com"
    }
  }
}
