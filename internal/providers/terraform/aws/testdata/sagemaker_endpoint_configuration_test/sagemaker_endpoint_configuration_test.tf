provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  skip_metadata_api_check     = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

# 1. Provisioned Instance Example
resource "aws_sagemaker_endpoint_configuration" "provisioned_config" {
  name = "provisioned-config"
  production_variants {
    variant_name           = "variant-1"
    model_name             = "my-model"
    instance_type          = "ml.m5.xlarge"
    initial_instance_count = 1
    volume_size_in_gb      = 20
  }
}

# 2. Serverless Example
resource "aws_sagemaker_endpoint_configuration" "serverless_config" {
  name = "serverless-config"
  production_variants {
    variant_name           = "serverless-variant"
    model_name             = "my-model"
    serverless_config {
      memory_size_in_mb = 2048
      max_concurrency   = 10
    }
  }
}

resource "aws_sagemaker_endpoint_configuration" "my_serverless_config" {
  name = "serverless-endpoint-config"

  production_variants {
    variant_name          = "AllTraffic"
    model_name            = "my-model-name"
    
    serverless_config {
      memory_size_in_mb       = 2048
      max_concurrency         = 10
      provisioned_concurrency = 2  # This triggers the "warm" costs
    }
  }
}
