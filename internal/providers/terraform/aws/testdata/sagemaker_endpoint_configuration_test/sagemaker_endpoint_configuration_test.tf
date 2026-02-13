provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  skip_metadata_api_check     = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_sagemaker_endpoint_configuration" "test" {
  name = "my-endpoint-config"
  production_variants {
    variant_name           = "variant-1"
    model_name             = "my-model"
    instance_type          = "ml.m5.xlarge"
    initial_instance_count = 2
    volume_size_in_gb      = 50
  }
}