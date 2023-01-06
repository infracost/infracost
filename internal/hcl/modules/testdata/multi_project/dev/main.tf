provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

module "local-module" {
  source        = "../modules/local-module"
  instance_type = "m5.4xlarge"
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

module "registry-module-different-name" {
  source  = "terraform-aws-modules/ec2-instance/aws"
  version = "3.4.0"

  name = "my-instance-2"

  ami                    = "ami-ebd02392"
  instance_type          = "t2.micro"
  key_name               = "user1"
  monitoring             = true
  vpc_security_group_ids = ["sg-12345678"]
  subnet_id              = "subnet-eddcdzz4"
}


module "git-module" {
  source = "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance.git"

  name = "my-instance"

  ami                    = "ami-ebd02392"
  instance_type          = "t2.micro"
  key_name               = "user1"
  monitoring             = true
  vpc_security_group_ids = ["sg-12345678"]
  subnet_id              = "subnet-eddcdzz4"
}

module "git-module-different-name" {
  source = "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance.git"

  name = "my-instance-2"

  ami                    = "ami-ebd02392"
  instance_type          = "t2.micro"
  key_name               = "user1"
  monitoring             = true
  vpc_security_group_ids = ["sg-12345678"]
  subnet_id              = "subnet-eddcdzz4"
}

module "another-git-module-only-in-dev" {
  source = "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance.git"

  name = "my-instance-2"

  ami                    = "ami-ebd02392"
  instance_type          = "t2.micro"
  key_name               = "user1"
  monitoring             = true
  vpc_security_group_ids = ["sg-12345678"]
  subnet_id              = "subnet-eddcdzz4"
}

