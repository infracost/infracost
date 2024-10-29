provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

locals {
  files = [
    { name = "cd", file = "instance.json", },
    { name = "sym", file = "sym-instance.json", },
    { name = "pd", file = "../../../testdata/instance.json", },
    {
      name = "abs",
      file = "/home/runner/work/infracost/infracost/cmd/infracost/testdata/breakdown_terraform_file_funcs/instance.json",
    },
    { name = "pdabs", file = "/home/runner/work/infracost/infracost/cmd/testdata/instance.json", },
  ]
  dirs = [
    { name = "cd", dir = "." },
    { name = "pd", dir = "../../../testdata" },
    { name = "abs", dir = "/home/runner/work/infracost/infracost/cmd/infracost/testdata/breakdown_terraform_file_funcs/" },
    { name = "pdabs", dir = "/home/runner/work/infracost/infracost/cmd/testdata/" },

  ]
  template_files = [
    { name = "cd", file = "templ.tftpl", },
    { name = "pd", file = "../../../testdata/templ.tftpl", },
    {
      name = "abs", file = "/home/runner/work/infracost/infracost/cmd/infracost/testdata/breakdown_terraform_file_funcs/templ.tftpl",
    },
    { name = "pdabs", file = "/home/runner/work/infracost/infracost/cmd/testdata/templ.tftpl", }
  ]
}

resource "aws_instance" "file" {
  for_each = { for f in local.files : f.name => f.file }

  ami           = "ami-674cbc1e"
  instance_type = jsondecode(file(each.value)).instance_type
}

module "mod_files" {
  source = "./modules/test"
}

resource "aws_instance" "fileexists" {
  for_each      = { for f in local.files : f.name => f.file }
  ami           = "ami-674cbc1e"
  instance_type = fileexists(each.value) ? "e" : "ne"
}

resource "aws_instance" "fileset" {
  for_each      = { for f in local.dirs : f.name => f.dir }
  ami           = "ami-674cbc1e"
  instance_type = length(fileset(each.value, "*.json"))
}

resource "aws_instance" "filemd5" {
  for_each      = { for f in local.files : f.name => f.file }
  ami           = "ami-674cbc1e"
  instance_type = filemd5(each.value)
}

resource "aws_instance" "template_file" {
  for_each      = { for f in local.template_files : f.name => f.file }
  ami           = "ami-674cbc1e"
  instance_type = jsondecode(templatefile(each.value, {})).instance_type
}
