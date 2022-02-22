variable "env_name" {
  description = "The name of the stack"
  type        = string
}

variable "eks_assume_role_policy" {
  description = "The Assume Role Policy for the EKS cluster"
  type        = string
}

variable "api_keys_table_arn" {
  description = "The ARN of the API keys table"
  type        = string
}

variable "database_instance_type" {
  description = "The instance type for the database"
  type        = string
}

variable "database_subnet_group_name" {
  description = "The name of the database subnet group"
  type        = string
}

variable "database_security_group_ids" {
  description = "The IDs of the database security groups"
  type        = list(string)
}

resource "aws_iam_role" "back_api_role" {
  name               = "${var.env_name}-back-api-role"
  description        = "Role for Back API"
  assume_role_policy = var.eks_assume_role_policy

  inline_policy {
    name   = "back-api-policy"
    policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [],
      "Resource": []
    }
  ]
}
    EOF
  }
}

resource "random_pet" "back_api_db_data" {
}

resource "aws_s3_bucket" "back_api_db_data" {
  bucket = "back-api-db-data-${random_pet.back_api_db_data.id}"
}

resource "random_password" "back_api_db_password" {
  length  = 32
  special = false
}

resource "aws_db_instance" "back_api_db" {
  identifier                = "${var.env_name}-back-api"
  allocated_storage         = 20
  max_allocated_storage     = 100
  engine                    = "mysql"
  instance_class            = var.database_instance_type
  multi_az                  = var.env_name == "prod"
  db_subnet_group_name      = var.database_subnet_group_name
  vpc_security_group_ids    = var.database_security_group_ids
  name                      = "backapi"
  username                  = "backapi"
  password                  = random_password.back_api_db_password.result
  backup_window             = "04:00-05:00"
  backup_retention_period   = 30
  maintenance_window        = "Sun:03:30-Sun:04:00"
  skip_final_snapshot       = var.env_name == "prod"
  final_snapshot_identifier = "back-api-db-final"
  parameter_group_name      = ""

  tags = {
    Environment = var.env_name
  }
}

output "back_api_db_address" {
  value = aws_db_instance.back_api_db.address
}

output "back_api_db_password" {
  value = random_password.back_api_db_password.result
}

output "role_arn" {
  value = aws_iam_role.back_api_role.arn
}

output "db_data_bucket_name" {
  value = aws_s3_bucket.back_api_db_data.bucket
}
