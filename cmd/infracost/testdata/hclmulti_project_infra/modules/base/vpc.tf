module "vpc" {
  source = "terraform-aws-modules/vpc/aws"

  name = var.env_name
  cidr = "10.0.0.0/16"

  azs              = [""]
  database_subnets = ["10.0.2.0/24"]
  private_subnets  = ["10.0.3.0/24"]
  public_subnets   = ["10.0.4.0/24"]

  enable_nat_gateway           = true
  single_nat_gateway           = true
  enable_dns_hostnames         = true
  create_database_subnet_group = true

  public_subnet_tags = {
    "kubernetes.io/cluster/${var.env_name}" = "shared"
    "kubernetes.io/role/elb"                = "1"
  }

  private_subnet_tags = {
    "kubernetes.io/cluster/${var.env_name}" = "shared"
    "kubernetes.io/role/internal-elb"       = "1"
  }

  tags = {
    Environment = var.env_name
  }
}

output "vpc_id" {
  value = module.vpc.vpc_id
}

output "database_subnet_group_name" {
  value = module.vpc.database_subnet_group_name
}

