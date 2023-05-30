provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_directory_service_directory" "simple_ad_small" {
  name     = "simplead-small-123456"
  type     = "SimpleAD"
  password = "SuperSecretPassw0rd"
  size     = "Small"

  vpc_settings {
    vpc_id     = "vpc-123456"
    subnet_ids = ["subnet-123456"]
  }
}

resource "aws_directory_service_directory" "simple_ad_large" {
  name     = "simplead-large-123456"
  type     = "SimpleAD"
  password = "SuperSecretPassw0rd"
  size     = "Large"

  vpc_settings {
    vpc_id     = "vpc-123456"
    subnet_ids = ["subnet-123456"]
  }
}

resource "aws_directory_service_directory" "ad_connector_small" {
  name     = "ad-connector-123456"
  password = "SuperSecretPassw0rd"
  size     = "Small"
  type     = "ADConnector"

  connect_settings {
    subnet_ids        = ["subnet_id"]
    vpc_id            = "vpc-123456"
    customer_dns_ips  = ["10.0.0.111"]
    customer_username = "username-123456"
  }
}

resource "aws_directory_service_directory" "microsoft_ad_standard" {
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  edition  = "Standard"
  type     = "MicrosoftAD"

  vpc_settings {
    vpc_id     = "vpc-123456"
    subnet_ids = ["subnet-123456"]
  }
}

resource "aws_directory_service_directory" "microsoft_ad_enterprise" {
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  edition  = "Enterprise"
  type     = "MicrosoftAD"

  vpc_settings {
    vpc_id     = "vpc-123456"
    subnet_ids = ["subnet-123456"]
  }
}

resource "aws_directory_service_directory" "simple_ad_with_usage" {
  name     = "simplead-small-123456"
  type     = "SimpleAD"
  password = "SuperSecretPassw0rd"
  size     = "Small"

  vpc_settings {
    vpc_id     = "vpc-123456"
    subnet_ids = ["subnet-123456"]
  }
}

resource "aws_directory_service_directory" "microsoft_ad_with_usage" {
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  edition  = "Standard"
  type     = "MicrosoftAD"

  vpc_settings {
    vpc_id     = "vpc-123456"
    subnet_ids = ["subnet-123456"]
  }
}
