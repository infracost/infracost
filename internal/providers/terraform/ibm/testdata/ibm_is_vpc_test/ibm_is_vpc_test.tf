
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

resource "ibm_is_vpc" "testVpc" {
  name = "test-vpc"
}

resource "ibm_is_vpc" "test_vpc_without_usage" {
  name = "test-vpc-without-usage"
}

