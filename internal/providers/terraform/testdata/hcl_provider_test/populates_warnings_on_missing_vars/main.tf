provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

variable "sensitive_test" {
  sensitive = true
}

variable "missing_var" {}
variable "test_token" {}
variable "token_test" {}
variable "test_secret" {}
variable "secret_test" {}
variable "test_password" {}
variable "password_test" {}
variable "test_username" {}
variable "username_test" {}
variable "test_api_key_test" {}
variable "test_expiration_date_test" {}
variable "AWS_ACCESS_KEY_ID" {}
variable "AWS_SECRET_ACCESS_KEY" {}
variable "access_key" {}
variable "secret_key" {}
variable "aws-secret-key" {}
variable "aws_profile" {}
variable "application-secrets" {}
variable "saml_role" {}
variable "_image" {}
variable "image_" {}

resource "aws_eip" "eip" {
  network_interface = "test"
}
