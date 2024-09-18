
terraform {
  required_providers {
    ibm = {
      source = "IBM-Cloud/ibm"
    }
  }
}

provider "ibm" {
  region = "us-south"
}

resource "ibm_is_vpc" "vpc" {
  name = "vpc"
}

resource "ibm_is_subnet" "subnet_1" {
  name            = "subnet-1"
  vpc             = ibm_is_vpc.vpc.id
  zone            = "us-south-1"
  ipv4_cidr_block = "10.240.0.0/24"
}

resource "ibm_is_subnet" "subnet_2" {
  name            = "subnet-2"
  vpc             = ibm_is_vpc.vpc.id
  zone            = "us-south-2"
  ipv4_cidr_block = "10.240.0.0/24"
}

resource "ibm_is_vpn_server" "vpn_server" {
  name            = "vpn-server"
  certificate_crn = ""
  client_ip_pool  = "10.0.0.0/20"
  subnets         = [ibm_is_subnet.subnet_1.id]
  client_authentication {
    method            = "username"
    identity_provider = "iam"
  }
}

resource "ibm_is_vpn_server" "vpn_server_without_usage" {
  name            = "vpn-server-no-usage"
  certificate_crn = ""
  client_ip_pool  = "10.0.0.0/20"
  subnets         = [ibm_is_subnet.subnet_2.id]
  client_authentication {
    method            = "username"
    identity_provider = "iam"
  }
}
