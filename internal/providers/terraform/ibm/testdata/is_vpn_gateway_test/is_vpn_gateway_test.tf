
terraform {
  required_providers {
    ibm = {
      source  = "IBM-Cloud/ibm"
      version = "1.63.0"
    }
  }
}

provider "ibm" {
  region     = "us-south"
  generation = 2
}

resource "ibm_is_vpc" "test_vpc" {
  name = "test-vpc"
}

resource "ibm_is_subnet" "test_vpc_subnet" {
  name            = "example-subnet"
  vpc             = ibm_is_vpc.test_vpc.id
  zone            = "us-south-1"
  ipv4_cidr_block = "10.240.0.0/24"
}

resource "ibm_is_subnet" "test_vpc_subnet2" {
  name            = "example-subnet2"
  vpc             = ibm_is_vpc.test_vpc.id
  zone            = "us-south-2"
  ipv4_cidr_block = "10.240.0.0/24"
}

resource "ibm_is_vpn_gateway" "test_vpc_vpn" {
  name   = "example-vpn-gateway"
  subnet = ibm_is_subnet.test_vpc_subnet.id
  mode   = "route"
}

resource "ibm_is_vpn_gateway" "test_vpc_vpn_without_usage" {
  name   = "example-vpn-gateway"
  subnet = ibm_is_subnet.test_vpc_subnet2.id
  mode   = "route"
}
