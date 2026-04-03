# AWS SES (Simple Email Service) Example
# 
# This example shows how to estimate costs for AWS SES resources using Infracost.
# SES is used for sending transactional and marketing emails at scale.
#
# Related Infracost Issue: https://github.com/infracost/infracost/issues/2998

provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  skip_metadata_api_check     = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

# SES Configuration Set for tracking email metrics
resource "aws_ses_configuration_set" "main" {
  name = "main-configuration-set"

  tracking_options {
    custom_redirect_domain = "tracking.example.com"
  }

  delivery_options {
    tls_policy = "Require"
  }

  reputation_metrics_enabled = true
  sending_enabled            = true
}

# SES Domain Identity
resource "aws_ses_domain_identity" "main" {
  domain = "example.com"
}

# SES Email Identity  
resource "aws_ses_email_identity" "notification" {
  email = "notifications@example.com"
}

# SES Email Identity for support
resource "aws_ses_email_identity" "support" {
  email = "support@example.com"
}

# SES Template for transactional emails
resource "aws_ses_template" "welcome" {
  name    = "welcome-email"
  subject = "Welcome to our service!"
  html    = <<EOF
<!DOCTYPE html>
<html>
<head>
  <title>Welcome</title>
</head>
<body>
  <h1>Welcome {{name}}!</h1>
  <p>Thank you for joining us.</p>
</body>
</html>
EOF
  text    = "Welcome {{name}}! Thank you for joining us."
}

# SES Template for password reset
resource "aws_ses_template" "password_reset" {
  name    = "password-reset"
  subject = "Password Reset Request"
  html    = <<EOF
<!DOCTYPE html>
<html>
<head>
  <title>Password Reset</title>
</head>
<body>
  <h1>Password Reset</h1>
  <p>Click <a href="{{link}}">here</a> to reset your password.</p>
</body>
</html>
EOF
  text    = "Click this link to reset your password: {{link}}"
}

# Optional: Dedicated IP Pool for high-volume sending
# Note: This has a monthly cost associated with it
# resource "aws_ses_dedicated_ip_pool" "main" {
#   pool_name = "main-pool"
# }

# Optional: SES Receipt Rule Set for inbound email
# resource "aws_ses_receipt_rule_set" "main" {
#   rule_set_name = "main-rule-set"
# }

output "configuration_set_name" {
  value = aws_ses_configuration_set.main.name
}

output "domain_identity_arn" {
  value = aws_ses_domain_identity.main.arn
}
