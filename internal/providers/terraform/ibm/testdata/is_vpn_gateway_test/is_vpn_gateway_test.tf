
terraform {
  required_providers {
    ibm = {
      source = "IBM-Cloud/ibm"
      version = "~> 1.40.0"
    }
  }
}

provider "ibm" {
    region = "us-south"
    generation = 2
}

resource "ibm_is_vpc" "testVpc" {
  name = "test-vpc"
}

resource "ibm_is_subnet" "testVpcSubnet" {
  name            = "example-subnet"
  vpc             = ibm_is_vpc.testVpc.id
  zone            = "us-south-1"
  ipv4_cidr_block = "10.240.0.0/24"
}

resource "ibm_is_vpn_gateway" "testVpcVpn" {
  name   = "example-vpn-gateway"
  subnet = ibm_is_subnet.testVpcSubnet.id
  mode   = "route"
}