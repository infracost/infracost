resource "random_pet" "web_bucket" {
  keepers = {
    # Generate a new pet name each time we switch to a new AMI id
    front_web_domain = var.front_web_domain
  }
}

resource "aws_s3_bucket" "front_web" {
  bucket = "${var.front_web_domain}-${random_pet.web_bucket.id}"
}

resource "aws_cloudfront_distribution" "front_web" {
  origin {
    custom_origin_config {
      http_port              = "80"
      https_port             = "443"
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["TLSv1", "TLSv1.1", "TLSv1.2"]
    }
    domain_name = aws_s3_bucket.front_web.website_endpoint
    origin_id   = var.front_web_domain
  }

  enabled             = true
  default_root_object = "index.html"

  default_cache_behavior {
    viewer_protocol_policy = "redirect-to-https"
    compress               = true
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = var.front_web_domain
    min_ttl                = 0
    default_ttl            = 86400
    max_ttl                = 31536000

    forwarded_values {
      query_string = false
      cookies {
        forward = "none"
      }
    }
  }

  custom_error_response {
    error_caching_min_ttl = 3000
    error_code            = 404
    response_code         = 200
    response_page_path    = "/index.html"
  }

  aliases = [var.front_web_domain]

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn = var.wildcard_cert_arn
    ssl_support_method  = "sni-only"
  }
}

resource "aws_route53_record" "front_web" {
  zone_id = "abc"
  name    = var.front_web_domain
  type    = "A"
  alias {
    name                   = aws_cloudfront_distribution.front_web.domain_name
    zone_id                = aws_cloudfront_distribution.front_web.hosted_zone_id
    evaluate_target_health = false
  }
}

output "front_web_bucket" {
  value = aws_s3_bucket.front_web.id
}
