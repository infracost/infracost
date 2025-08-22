provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

# Add example resources for CloudfrontFunction below

# resource "aws_cloudfront_function" "cloudfront_function" {
# }
resource "aws_cloudfront_function" "bibit_deeplink_cf_function" {
  name    = "test_func"
  runtime = "cloudfront-js-2.0"
  comment = "Bibit Deeplink CF Function Script"
  publish = true
  code    = <<-EOT
  async function handler(event) {
      var request = event.request;
      var uri = request.uri;
      
      // Check whether the URI is missing a file name.
      if (uri.endsWith('/')) {
          request.uri += 'index.html';
      } 
      // Check whether the URI is missing a file extension.
      else if (!uri.includes('.')) {
          request.uri += '/index.html';
      }

      return request;
  }
  EOT
}
