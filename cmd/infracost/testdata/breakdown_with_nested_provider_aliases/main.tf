provider "aws" {
  alias  = "lookup"
  region = "eu-west-2"
}

module "account_lookup" {
  source = "./account_lookup"
  providers = {
    aws = aws.lookup
  }
}

provider "aws" {
  region = module.account_lookup.data.region_name
  assume_role {
    role_arn = module.account_lookup.data.role
  }
}

resource "aws_instance" "web" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t2.micro"
}
