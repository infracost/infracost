provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_sagemaker_endpoint_configuration" "instance_config" {
  name = "instance-config"
  production_variants {
    variant_name           = "instance-config-variant"
    model_name             = "my-model"
    instance_type          = "ml.m5.xlarge"
    initial_instance_count = 1
    volume_size_in_gb      = 20
  }

  production_variants {
    variant_name           = "instance-config-variant2"
    model_name             = "my-model"
    instance_type          = "ml.m5.large"
    initial_instance_count = 1
    volume_size_in_gb      = 20
  }
}

# 2. Serverless Example
resource "aws_sagemaker_endpoint_configuration" "serverless_config" {
  name = "serverless-config"
  production_variants {
    variant_name = "serverless-variant"
    model_name   = "my-model"
    serverless_config {
      memory_size_in_mb = 2048
      max_concurrency   = 10
    }
  }
}

resource "aws_sagemaker_endpoint_configuration" "serverless_config_multiple_variants" {
  name = "serverless-config-multiple-variants"
  production_variants {
    variant_name = "serverless-variant"
    model_name   = "my-model"
    serverless_config {
      memory_size_in_mb = 2048
      max_concurrency   = 10
    }
  }

  production_variants {
    variant_name = "serverless-variant2"
    model_name   = "my-model"
    serverless_config {
      memory_size_in_mb = 1024
      max_concurrency   = 10
    }
  }
}


resource "aws_sagemaker_endpoint_configuration" "serverless_config_provisioned_concurrency" {
  name = "serverless-config-provisioned-concurrency"

  production_variants {
    variant_name = "serverless-config-provisioned-concurrency-variant"
    model_name   = "my-model"

    serverless_config {
      memory_size_in_mb       = 2048
      max_concurrency         = 10
      provisioned_concurrency = 2
    }
  }
}

resource "aws_sagemaker_endpoint_configuration" "shadow_test_config" {
  name = "shadow-test-config"
  production_variants {
    variant_name           = "Primary-Live-variant"
    model_name             = "my-old-model"
    instance_type          = "ml.m5.large"
    initial_instance_count = 1
  }

  # Shadow Variant (Billing is identical to production)
  shadow_production_variants {
    variant_name           = "Shadow-variant"
    model_name             = "my-new-model"
    instance_type          = "ml.m5.xlarge" # testing a larger instance
    initial_instance_count = 1
  }
}
