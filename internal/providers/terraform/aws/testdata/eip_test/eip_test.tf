provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_eip" "eip1" {}

resource "aws_eip" "eip2" {}

resource "aws_nat_gateway" "nat_gw" {
  allocation_id = aws_eip.eip2.id
  subnet_id     = "subnet-12345678"
}

resource "aws_eip" "eip3" {}
resource "aws_instance" "instance" {
  ami           = "fake_ami"
  instance_type = "t3.medium"
}

resource "aws_eip_association" "assoc" {
  instance_id   = aws_instance.instance.id
  allocation_id = aws_eip.eip3.id
}

resource "aws_eip" "eip4" {}
resource "aws_lb" "example" {
  name               = "example"
  load_balancer_type = "network"

  subnet_mapping {
    subnet_id     = "subnet-12345678"
    allocation_id = aws_eip.eip3.id
  }

  subnet_mapping {
    subnet_id     = "subnet-12345679"
    allocation_id = aws_eip.eip4.id
  }
}
