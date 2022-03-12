variable "env_name" {
  description = "The name of the stack"
  type        = string
}

variable "vpc_id" {
  description = "The VPC ID"
  type        = string
}

variable "eks_assume_role_policy" {
  description = "The Assume Role Policy for the EKS cluster"
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

variable "root_domain" {
  description = "The root domain of the hosted zone"
  type        = string
}

variable "api_keys_table_arn" {
  description = "The ARN of the API keys table"
  type        = string
}

variable "wildcard_cert_arn" {
  description = "The ARN of the wildcard cert"
  type        = string
}

variable "front_web_domain" {
  description = "The domain for the front web UI"
  type        = string
}
