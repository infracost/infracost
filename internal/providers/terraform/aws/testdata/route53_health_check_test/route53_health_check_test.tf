provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_route53_health_check" "simple" {
  failure_threshold = "5"
  fqdn              = "example.com"
  port              = 80
  request_interval  = "30"
  resource_path     = "/"
  type              = "HTTP"
}

resource "aws_route53_health_check" "https" {
  failure_threshold = "5"
  fqdn              = "example.com"
  port              = 443
  request_interval  = "30"
  resource_path     = "/"
  type              = "HTTPS"
}

resource "aws_route53_health_check" "ssl_string_match" {
  failure_threshold = "5"
  fqdn              = "example.com"
  port              = 443
  request_interval  = "30"
  resource_path     = "/"
  type              = "HTTPS_STR_MATCH"
  search_string     = "TestingTesting"
}

resource "aws_route53_health_check" "ssl_latency_string_match" {
  failure_threshold = "5"
  fqdn              = "example.com"
  port              = 443
  request_interval  = "30"
  resource_path     = "/"
  type              = "HTTPS_STR_MATCH"
  search_string     = "TestingTesting"
  measure_latency   = true
}

resource "aws_route53_health_check" "ssl_latency_interval_string_match" {
  failure_threshold = "5"
  fqdn              = "example.com"
  port              = 443
  request_interval  = "10"
  resource_path     = "/"
  type              = "HTTPS_STR_MATCH"
  search_string     = "TestingTesting"
  measure_latency   = true
}

resource "aws_route53_health_check" "simple_withUsage" {
  failure_threshold = "5"
  fqdn              = "example.com"
  port              = 80
  request_interval  = "30"
  resource_path     = "/"
  type              = "HTTP"
}

resource "aws_route53_health_check" "https_withUsage" {
  failure_threshold = "5"
  fqdn              = "example.com"
  port              = 443
  request_interval  = "30"
  resource_path     = "/"
  type              = "HTTPS"
}

resource "aws_route53_health_check" "ssl_string_match_withUsage" {
  failure_threshold = "5"
  fqdn              = "example.com"
  port              = 443
  request_interval  = "30"
  resource_path     = "/"
  type              = "HTTPS_STR_MATCH"
  search_string     = "TestingTesting"
}

resource "aws_route53_health_check" "ssl_latency_string_match_withUsage" {
  failure_threshold = "5"
  fqdn              = "example.com"
  port              = 443
  request_interval  = "30"
  resource_path     = "/"
  type              = "HTTPS_STR_MATCH"
  search_string     = "TestingTesting"
  measure_latency   = true
}

resource "aws_route53_health_check" "ssl_latency_interval_string_match_withUsage" {
  failure_threshold = "5"
  fqdn              = "example.com"
  port              = 443
  request_interval  = "10"
  resource_path     = "/"
  type              = "HTTPS_STR_MATCH"
  search_string     = "TestingTesting"
  measure_latency   = true
}
