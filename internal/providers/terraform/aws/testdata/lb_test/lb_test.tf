provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_lb" "lb1" {
  load_balancer_type = "application"
  subnets            = ["subnet-12345678"]
}

resource "aws_alb" "alb1" {
  subnets = ["subnet-12345678"]
}

resource "aws_lb" "nlb1" {
  load_balancer_type = "network"
  subnets            = ["subnet-12345678"]
}

resource "aws_lb" "alb1_usage" {
  load_balancer_type = "application"
  subnets            = ["subnet-12345678"]
}

resource "aws_lb" "nlb1_usage" {
  load_balancer_type = "network"
  subnets            = ["subnet-12345678"]
}
