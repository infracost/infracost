terraform {
  required_providers {
    ibm = {
      source = "IBM-Cloud/ibm"
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

resource "ibm_is_public_gateway" "example" {
  name = "example-gateway"
  vpc  = ibm_is_vpc.test_vpc.id
  zone = "us-south-1"
}