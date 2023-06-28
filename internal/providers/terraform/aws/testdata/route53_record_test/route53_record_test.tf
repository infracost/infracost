provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_route53_zone" "zone1" {
  name = "example.com"
}

resource "aws_route53_record" "standard" {
  zone_id = aws_route53_zone.zone1.zone_id
  name    = "standard.example.com"
  type    = "A"
  ttl     = "300"
  records = ["10.0.0.1"]
}

resource "aws_route53_zone" "zone_withUsage" {
  name = "example.com"
}

resource "aws_route53_record" "my_record_withUsage" {
  zone_id = aws_route53_zone.zone_withUsage.zone_id
  name    = "standard.example.com"
  type    = "A"
  ttl     = "300"
  records = ["10.0.0.1"]
}
