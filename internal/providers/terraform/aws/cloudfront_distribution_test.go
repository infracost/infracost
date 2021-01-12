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
					Name: "Regional data transfer out to origin",
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
							Name:      "Hong Kong, Philippines, Asia Pacific",
							PriceHash: "63a411ecbb6d084a9e9c15b49c4a3ec9-b1ae3861dc57e2db217fa83a7420374f",
						},
						{
							Name:      "India",
							PriceHash: "74d31f8195b5487364d2ae10b0b144c4-b1ae3861dc57e2db217fa83a7420374f",
						},
					},
				},
				{
					Name: "Request pricing for all http methods",
					SubResourceChecks: []testutil.ResourceCheck{
						{
							Name: "HTTP requests",
							CostComponentChecks: []testutil.CostComponentCheck{
								{
									Name:      "United States, Mexico, & Canada",
									PriceHash: "6e7bb9693c7bdc3c1b09a5ad0cd11a4a-4a9dfd3965ffcbab75845ead7a27fd47",
								},
								{
									Name:      "Europe & Israel",
									PriceHash: "f81d8aa74fae2d32a4149a85920f3255-4a9dfd3965ffcbab75845ead7a27fd47",
								},
								{
									Name:      "South Africa, Kenya, & Middle East",
									PriceHash: "c64d2813fa3777ace1a1006389217239-4a9dfd3965ffcbab75845ead7a27fd47",
								},
								{
									Name:      "South America",
									PriceHash: "f0243692bd53ed2cef6ed6445b0c5683-4a9dfd3965ffcbab75845ead7a27fd47",
								},
								{
									Name:      "Japan",
									PriceHash: "681d410b9400be8fb5e7e2d1b089d159-4a9dfd3965ffcbab75845ead7a27fd47",
								},
								{
									Name:      "Australia & New Zealand",
									PriceHash: "4e86dc6c95675a4c8dd4ac876a30ab3c-4a9dfd3965ffcbab75845ead7a27fd47",
								},
								{
									Name:      "Hong Kong, Philippines, Asia Pacific",
									PriceHash: "871d73c17fc8c93de0ccdbc2c9c470d7-4a9dfd3965ffcbab75845ead7a27fd47",
								},
								{
									Name:      "India",
									PriceHash: "2632f4cda76bc34285fb6cd5fb894ee4-4a9dfd3965ffcbab75845ead7a27fd47",
								},
							},
						},
						{
							Name: "HTTPS requests",
							CostComponentChecks: []testutil.CostComponentCheck{
								{
									Name:      "United States, Mexico, & Canada",
									PriceHash: "8890fabb60883960c9178fe0e753e47e-4a9dfd3965ffcbab75845ead7a27fd47",
								},
								{
									Name:      "Europe & Israel",
									PriceHash: "63c72b02594fc500d149b54e4248e38b-4a9dfd3965ffcbab75845ead7a27fd47",
								},
								{
									Name:      "South Africa, Kenya, & Middle East",
									PriceHash: "a1527c0b56940465cf2a5eabf97e45f0-4a9dfd3965ffcbab75845ead7a27fd47",
								},
								{
									Name:      "South America",
									PriceHash: "3388ba97d6c8373e5c6de6ff51b431af-4a9dfd3965ffcbab75845ead7a27fd47",
								},
								{
									Name:      "Japan",
									PriceHash: "3f75cf910bfbe3e47bbff04ed01e3986-4a9dfd3965ffcbab75845ead7a27fd47",
								},
								{
									Name:      "Australia & New Zealand",
									PriceHash: "358f87101e7deff58a09cc76e1de7bd3-4a9dfd3965ffcbab75845ead7a27fd47",
								},
								{
									Name:      "Hong Kong, Philippines, Asia Pacific",
									PriceHash: "1931ee7f0715a77116c6c4a7e1eecf49-4a9dfd3965ffcbab75845ead7a27fd47",
								},
								{
									Name:      "India",
									PriceHash: "0a703a33e830797459e6a0226336bb19-4a9dfd3965ffcbab75845ead7a27fd47",
								},
							},
						},
					},
				},
				{
					Name: "Origin shield request pricing for all http methods",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:      "United States",
							PriceHash: "9a59a3308256aab9256b6a421fd072d9-4a9dfd3965ffcbab75845ead7a27fd47",
						},
						{
							Name:      "Europe",
							PriceHash: "43f5e56d0b879abe92fc71f280d995fc-4a9dfd3965ffcbab75845ead7a27fd47",
						},
						{
							Name:      "South America",
							PriceHash: "224f2fff366333b0e6dfeb454010be9f-4a9dfd3965ffcbab75845ead7a27fd47",
						},
						{
							Name:      "Japan",
							PriceHash: "1169ba622705234fd01b29ed53173f2d-4a9dfd3965ffcbab75845ead7a27fd47",
						},
						{
							Name:      "Australia",
							PriceHash: "57674bc88879a321596331ff12c624fa-4a9dfd3965ffcbab75845ead7a27fd47",
						},
						{
							Name:      "Singapore",
							PriceHash: "57e69a82635268b50499099c6311b694-4a9dfd3965ffcbab75845ead7a27fd47",
						},
						{
							Name:      "South Korea",
							PriceHash: "f1f36dcbd00e0b5a78dd8134b1314350-4a9dfd3965ffcbab75845ead7a27fd47",
						},
						{
							Name:      "India",
							PriceHash: "dce9a91d009b3e40ab41d992d6009779-4a9dfd3965ffcbab75845ead7a27fd47",
						},
					},
				},
			},
		},
		{
			Name:      "aws_s3_bucket.b",
			SkipCheck: true,
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}
