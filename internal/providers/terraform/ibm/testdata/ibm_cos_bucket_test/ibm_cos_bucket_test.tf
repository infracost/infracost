terraform {
  required_providers {
    ibm = {
      source = "IBM-Cloud/ibm"
      version = "~> 1.12.0"
    }
  }
}

provider "ibm" {
    region = "global"
}

resource "ibm_resource_group" "cos_group" {
  name = "cos-resource-group"
}

resource "ibm_resource_instance" "cos_instance" {
  name              = "cos-instance"
  resource_group_id = ibm_resource_group.cos_group.id
  service           = "cloud-object-storage"
  plan              = "standard"
  location          = "us-south"
}

resource "ibm_cos_bucket" "standard-us-south" {
  bucket_name          = "a-standard-bucket-at-us-south"
  resource_instance_id = ibm_resource_instance.cos_instance.id
  storage_class        = "standard"
  region_location = "us-south"
}
