variable "subnet_id" {
  type = string
}

resource "aws_network_interface" "submodule_eip_network_interface" {
  subnet_id   = var.subnet_id
  private_ips = ["10.0.0.1"]
}

resource "aws_eip" "submodule_nat_eip" {
  network_interface = aws_network_interface.submodule_eip_network_interface.id
  vpc = true
}

resource "aws_nat_gateway" "submodule_nat" {
  subnet_id     = var.subnet_id
  allocation_id = aws_eip.submodule_nat_eip.id
}

