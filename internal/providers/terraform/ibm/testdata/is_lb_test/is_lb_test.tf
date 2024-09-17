terraform {
  required_providers {
    ibm = {
      source  = "IBM-Cloud/ibm"
      version = "1.63.0"
    }
  }
}

provider "ibm" {
  region = "us-south"
}

resource "ibm_is_vpc" "test_vpc" {
  name = "test-vpc"
}


resource "ibm_is_subnet" "test_subnet" {
  name            = "example-subnet"
  vpc             = ibm_is_vpc.test_vpc.id
  zone            = "us-south-1"
  ipv4_cidr_block = "10.240.0.0/24"
}

resource "ibm_is_lb" "network_lb" {
  name    = "example-load-balancer"
  subnets = [ibm_is_subnet.test_subnet.id]
  profile = "network-fixed"
}

resource "ibm_is_lb" "application_lb" {
  name    = "example-load-balancer"
  subnets = [ibm_is_subnet.test_subnet.id]
}

resource "ibm_is_lb" "application_lb_without_usage" {
  name    = "example-load-balancer"
  subnets = [ibm_is_subnet.test_subnet.id]
}
