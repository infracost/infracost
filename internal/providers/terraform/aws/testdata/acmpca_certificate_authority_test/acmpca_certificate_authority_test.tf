provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}


resource "aws_acmpca_certificate_authority" "private_ca_noUsage" {
  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"
    subject {
      common_name = "private-ca.com"
    }
  }
}

resource "aws_acmpca_certificate_authority" "private_ca" {
  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"
    subject {
      common_name = "private-ca.com"
    }
  }
}

resource "aws_acmpca_certificate_authority" "private_ca_tiered" {
  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"
    subject {
      common_name = "private-ca.com"
    }
  }
}

resource "aws_acmpca_certificate_authority" "private_ca_more_tiered" {
  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"
    subject {
      common_name = "private-ca.com"
    }
  }
}

resource "aws_acmpca_certificate_authority" "short_lived_ca_noUsage" {
  usage_mode = "SHORT_LIVED_CERTIFICATE"
  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"
    subject {
      common_name = "short-lived-ca.com"
    }
  }
}

resource "aws_acmpca_certificate_authority" "short_lived_ca" {
  usage_mode = "SHORT_LIVED_CERTIFICATE"
  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"
    subject {
      common_name = "short-lived-ca.com"
    }
  }
}
