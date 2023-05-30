provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_route53_resolver_endpoint" "test" {
  direction          = "INBOUND"
  security_group_ids = ["sg-1233456"]

  ip_address {
    subnet_id = "subnet-123456"
  }

  ip_address {
    subnet_id = "subnet-654321"
  }
}

resource "aws_route53_resolver_endpoint" "test_withUsage1B" {
  direction          = "INBOUND"
  security_group_ids = ["sg-1233456"]

  ip_address {
    subnet_id = "subnet-123456"
  }

  ip_address {
    subnet_id = "subnet-654321"
  }
}

resource "aws_route53_resolver_endpoint" "test_withUsage2B" {
  direction          = "INBOUND"
  security_group_ids = ["sg-1233456"]

  ip_address {
    subnet_id = "subnet-123456"
  }

  ip_address {
    subnet_id = "subnet-654321"
  }
}

