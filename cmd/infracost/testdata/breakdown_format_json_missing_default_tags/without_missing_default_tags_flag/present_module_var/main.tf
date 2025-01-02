module "mymod" {
  source     = "./module"
  aws_region = "us-east-1"
  tags = {
    T1 = "V1"
  }
}
