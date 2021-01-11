package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestCloudfrontDistribution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_s3_bucket" "b" {
			bucket = "mybucket"
			acl    = "private"
		
			tags = {
			Name = "My bucket"
			}
		}
		
		locals {
			s3_origin_id = "myS3Origin"
		}
		
		resource "aws_cloudfront_distribution" "s3_distribution" {
			origin {
			domain_name = aws_s3_bucket.b.bucket_regional_domain_name
			origin_id   = local.s3_origin_id
		
			s3_origin_config {
				origin_access_identity = "origin-access-identity/cloudfront/ABCDEFG1234567"
			}
			}
		
			enabled             = true
			is_ipv6_enabled     = true
			comment             = "Some comment"
			default_root_object = "index.html"
		
			logging_config {
			include_cookies = false
			bucket          = "mylogs.s3.amazonaws.com"
			prefix          = "myprefix"
			}
		
			aliases = ["mysite.example.com", "yoursite.example.com"]
		
			default_cache_behavior {
			allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
			cached_methods   = ["GET", "HEAD"]
			target_origin_id = local.s3_origin_id
		
			forwarded_values {
				query_string = false
		
				cookies {
				forward = "none"
				}
			}
		
			viewer_protocol_policy = "allow-all"
			min_ttl                = 0
			default_ttl            = 3600
			max_ttl                = 86400
			}
		
			# Cache behavior with precedence 0
			ordered_cache_behavior {
			path_pattern     = "/content/immutable/*"
			allowed_methods  = ["GET", "HEAD", "OPTIONS"]
			cached_methods   = ["GET", "HEAD", "OPTIONS"]
			target_origin_id = local.s3_origin_id
		
			forwarded_values {
				query_string = false
				headers      = ["Origin"]
		
				cookies {
				forward = "none"
				}
			}
		
			min_ttl                = 0
			default_ttl            = 86400
			max_ttl                = 31536000
			compress               = true
			viewer_protocol_policy = "redirect-to-https"
			}
		
			# Cache behavior with precedence 1
			ordered_cache_behavior {
			path_pattern     = "/content/*"
			allowed_methods  = ["GET", "HEAD", "OPTIONS"]
			cached_methods   = ["GET", "HEAD"]
			target_origin_id = local.s3_origin_id
		
			forwarded_values {
				query_string = false
		
				cookies {
				forward = "none"
				}
			}
		
			min_ttl                = 0
			default_ttl            = 3600
			max_ttl                = 86400
			compress               = true
			viewer_protocol_policy = "redirect-to-https"
			}
		
			price_class = "PriceClass_200"
		
			restrictions {
			geo_restriction {
				restriction_type = "whitelist"
				locations        = ["US", "CA", "GB", "DE"]
			}
			}
		
			tags = {
			Environment = "production"
			}
		
			viewer_certificate {
			cloudfront_default_certificate = true
			}
		}
	`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_cloudfront_distribution.s3_distribution",
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "Regional Data Transfer Out to Origin",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:      "United States, Mexico, & Canada",
							PriceHash: "0c8dbb9a1aad0159dba32a7dcd48b384-b1ae3861dc57e2db217fa83a7420374f",
						},
						{
							Name:      "Europe & Israel",
							PriceHash: "afb13cd55f419b70212c5767ff502944-b1ae3861dc57e2db217fa83a7420374f",
						},
						{
							Name:      "South Africa, Kenya, & Middle East",
							PriceHash: "7cbab97f2b54211d7654b0e4ba3f3c70-b1ae3861dc57e2db217fa83a7420374f",
						},
						{
							Name:      "South America",
							PriceHash: "5cc794b11c9e61704a9dfdeaa95721d6-b1ae3861dc57e2db217fa83a7420374f",
						},
						{
							Name:      "Japan",
							PriceHash: "5456abd68dfb61de5a60286196e52205-b1ae3861dc57e2db217fa83a7420374f",
						},
						{
							Name:      "Australia & New Zealand",
							PriceHash: "80125f460392b4b600eb5954de37e913-b1ae3861dc57e2db217fa83a7420374f",
						},
						{
							Name:      "Hong Kong, Philippines, Singapore, South Korea, Taiwan, & Thailand",
							PriceHash: "63a411ecbb6d084a9e9c15b49c4a3ec9-b1ae3861dc57e2db217fa83a7420374f",
						},
						{
							Name:      "India",
							PriceHash: "74d31f8195b5487364d2ae10b0b144c4-b1ae3861dc57e2db217fa83a7420374f",
						},
					},
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}
