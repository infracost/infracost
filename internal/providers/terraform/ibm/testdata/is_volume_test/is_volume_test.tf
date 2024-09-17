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

resource "ibm_is_volume" "custom_volume" {
  name     = "example-volume"
  profile  = "custom"
  zone     = "us-south-1"
  iops     = 1000
  capacity = 200
}

resource "ibm_is_volume" "general_purpose_volume" {
  name     = "example-volume"
  profile  = "general-purpose"
  zone     = "us-south-1"
  capacity = 200
}

resource "ibm_is_volume" "general_purpose_5iops" {
  name     = "example-volume"
  profile  = "5iops-tier"
  zone     = "us-south-1"
  capacity = 400
}

resource "ibm_is_volume" "general_purpose_volume2" {
  name     = "example-volume"
  profile  = "general-purpose"
  zone     = "us-south-1"
  capacity = 200
}
