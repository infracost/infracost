provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_api_gateway_stage" "cache_1" {
  rest_api_id           = "api-id-1"
  stage_name            = "cache-stage-1"
  deployment_id         = "deployment-id-1"
  cache_cluster_enabled = true
  cache_cluster_size    = 0.5
}

resource "aws_api_gateway_stage" "cache_2" {
  rest_api_id           = "api-id-2"
  stage_name            = "cache-stage-2"
  deployment_id         = "deployment-id-2"
  cache_cluster_enabled = true
  cache_cluster_size    = 237
}

resource "aws_api_gateway_stage" "default_cache" {
  rest_api_id   = "api-id-3"
  stage_name    = "cache-stage-3"
  deployment_id = "deployment-id-3"
}

resource "aws_api_gateway_stage" "disabled_cache" {
  rest_api_id           = "api-id-4"
  stage_name            = "cache-stage-4"
  deployment_id         = "deployment-id-4"
  cache_cluster_enabled = false
  cache_cluster_size    = 237
}
