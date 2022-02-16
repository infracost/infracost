variable "instance_type" {
}

resource "aws_instance" "web_app" {
  ami           = "ami-674cbc1e"
  instance_type = var.instance_type
}

module "nested-local-module" {
  source        = "./nested-local-module"
  instance_type = "m5.8xlarge"
}

module "nested-registry-module" {
  source  = "terraform-aws-modules/sns/aws"
  version = "3.1.0"

  name = "my-topic"
}

module "nested-git-module" {
  source = "git::https://github.com/terraform-aws-modules/terraform-aws-sns.git"

  name = "my-topic"
}
