
terraform {
  required_providers {
    ibm = {
      source = "IBM-Cloud/ibm"
      version = "~> 1.12.0"
    }
  }
}

provider "ibm" {
    region = "us-south"
}

resource "ibm_is_vpc" "testVpc" {
  name = "test-vpc"
}

resource "ibm_is_subnet" "testSubnet" {
  name            = "test-subnet"
  vpc             = ibm_is_vpc.testVpc.id
  zone            = "us-south-1"
  ipv4_cidr_block = "10.240.0.0/24"
}

resource "ibm_is_ssh_key" "testSshKey" {
  name       = "test-ssh"
  public_key = file("~/.ssh/id_rsa.pub")
}


resource "ibm_is_instance" "testInstance" {
  name    = "test-instance-1"
  image   = "r006-f137ea64-0d27-4d81-afe0-353fd0557e81"
  profile = "cx2-2x4"

  primary_network_interface {
    subnet = ibm_is_subnet.testSubnet.id
  }

  network_interfaces {
    name   = "eth1"
    subnet = ibm_is_subnet.testSubnet.id
  }

  vpc  = ibm_is_vpc.testVpc.id
  zone = "us-south-1"
  keys = [ibm_is_ssh_key.testSshKey.id]
}
