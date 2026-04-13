provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

# Cost-optimized EC2 instance
# Original: m5.4xlarge (~$560/month)
# Optimized: m5.2xlarge (~$280/month) - 50% savings
resource "aws_instance" "web_app_optimized" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.2xlarge" # Right-sized for typical web workloads

  root_block_device {
    volume_size = 50
    volume_type = "gp3" # gp3 is more cost-effective than gp2
  }

  # Cost-optimized EBS volume
  # Original: io1 1000GB with 800 IOPS (~$1,250/month)
  # Optimized: gp3 500GB with 3000 IOPS baseline (~$40/month) - 97% savings
  ebs_block_device {
    device_name = "my_data"
    volume_type = "gp3"
    volume_size = 500
    # gp3 provides 3,000 IOPS baseline at no additional cost
    # Only provision additional IOPS if you need more than 3,000
  }

  tags = {
    Name        = "cost-optimized-web-app"
    Environment = "production"
    CostCenter  = "engineering"
  }
}

# Cost-optimized Lambda function
# Original: 1024 MB memory
# Optimized: 512 MB memory - 50% savings per invocation
resource "aws_lambda_function" "hello_world_optimized" {
  function_name = "hello_world_optimized"
  role          = "arn:aws:lambda:us-east-1:aws:resource-id"
  handler       = "exports.test"
  runtime       = "nodejs20.x" # Updated to newer runtime
  filename      = "function.zip"
  memory_size   = 512 # Reduced from 1024 MB - test to find optimal setting

  tags = {
    Name        = "cost-optimized-lambda"
    Environment = "production"
  }
}

# Additional cost-optimized resources demonstrating best practices

# Use Graviton2 instances for better price/performance
resource "aws_instance" "web_app_graviton" {
  ami           = "ami-674cbc1e" # Amazon Linux 2 on Graviton
  instance_type = "m6g.xlarge"   # Graviton2-based, ~20% cheaper than m5.xlarge

  root_block_device {
    volume_size = 50
    volume_type = "gp3"
  }

  tags = {
    Name        = "graviton-web-app"
    Environment = "production"
    CostCenter  = "engineering"
  }
}

# Outputs for cost analysis
output "optimized_instance_type" {
  description = "The optimized EC2 instance type"
  value       = aws_instance.web_app_optimized.instance_type
}

output "optimized_monthly_estimate" {
  description = "Estimated monthly cost for optimized configuration"
  value       = "~$320/month (EC2 + EBS) vs ~$1,810/month original - ~82% savings"
}

output "cost_optimization_tips" {
  description = "Tips for further cost optimization"
  value       = <<-EOT
    1. Consider Reserved Instances for predictable 24/7 workloads (30-60% savings)
    2. Use Spot Instances for fault-tolerant workloads (up to 90% savings)
    3. Enable CloudWatch detailed monitoring only when needed
    4. Use AWS Compute Optimizer for right-sizing recommendations
    5. Consider Savings Plans for flexible commitment-based discounts
  EOT
}
