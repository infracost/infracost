provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

variable "missing_var" {
  type = bool
}

variable "test_token" {
  type = bool
}

variable "token_test" {
  type = bool
}

variable "test_secret" {
  type = bool
}

variable "secret_test" {
  type = bool
}

variable "test_password" {
  type = bool
}

variable "password_test" {
  type = bool
}

variable "test_username" {
  type = bool
}

variable "username_test" {
  type = bool
}

variable "test_api_key_test" {
  type = bool
}

variable "test_expiration_date_test" {
  type = bool
}

variable "sensitive_test" {
  type      = bool
  sensitive = true
}

resource "aws_eip" "eip" {
  network_interface = "test"
}
