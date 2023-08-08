terraform {
  source = "git::https://github.com/terraform-aws-modules/terraform-aws-eks.git//modules/fargate-profile?ref=v19.15.2"
}

inputs = {
  name         = "separate-fargate-profile"
  cluster_name = "my-cluster"

  subnet_ids = ["subnet-abcde012", "subnet-bcde012a", "subnet-fghi345a"]
  selectors = [{
    namespace = "kube-system"
  }]

  tags = {
    Environment = "dev"
    Terraform   = "true"
  }
}
