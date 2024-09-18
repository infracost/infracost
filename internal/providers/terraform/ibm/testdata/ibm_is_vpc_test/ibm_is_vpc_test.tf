
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

resource "ibm_is_vpc" "testVpc" {
  name = "test-vpc"
}

resource "ibm_is_vpc" "test_vpc_without_usage" {
  name = "test-vpc-without-usage"
}

