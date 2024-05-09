locals {
  files = [
    { name = "cd", file = "${path.module}/instance.json", },
    { name = "rootcd", file = "${path.module}/../../instance.json", },
    { name = "pd", file = "${path.module}/../../../../testdata/instance.json", },
  ]
}

resource "aws_instance" "file" {
  for_each = { for f in local.files : f.name => f.file }

  ami           = "ami-674cbc1e"
  instance_type = jsondecode(file(each.value)).instance_type
}

resource "aws_instance" "fileexists" {
  for_each = { for f in local.files : f.name => f.file }

  ami           = "ami-674cbc1e"
  instance_type = fileexists(each.value) ? "e" : "ne"
}
