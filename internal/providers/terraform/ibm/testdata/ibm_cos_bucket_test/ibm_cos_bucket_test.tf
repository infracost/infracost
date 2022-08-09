terraform {
  required_providers {
    ibm = {
      source  = "IBM-Cloud/ibm"
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
  bucket_name          = "standard-bucket-at-us-south"
  resource_instance_id = ibm_resource_instance.cos_instance.id
  storage_class        = "standard"
  region_location      = "us-south"
}

resource "ibm_cos_bucket" "smart-us-south" {
  bucket_name          = "smart-bucket-at-us-south"
  resource_instance_id = ibm_resource_instance.cos_instance.id
  storage_class        = "smart"
  region_location      = "us-south"
}

resource "ibm_cos_bucket" "aspera-us" {
  bucket_name           = "aspera-bucket-at-us"
  resource_instance_id  = ibm_resource_instance.cos_instance.id
  storage_class         = "standard"
  cross_region_location = "us"
}

resource "ibm_cos_bucket" "archive-us-south" {
  bucket_name          = "archive-bucket-at-us-south"
  resource_instance_id = ibm_resource_instance.cos_instance.id
  storage_class        = "standard"
  region_location      = "us-south"
}

resource "ibm_cos_bucket" "standard-ams03" {
  bucket_name          = "standard-bucket-at-ams03"
  resource_instance_id = ibm_resource_instance.cos_instance.id
  storage_class        = "standard"
  single_site_location      = "ams03"
}