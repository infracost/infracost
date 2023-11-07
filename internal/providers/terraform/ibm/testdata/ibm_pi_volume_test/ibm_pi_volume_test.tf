terraform {
  required_providers {
    ibm = {
      source  = "IBM-Cloud/ibm"
      version = "1.58.0"
    }
  }
}

provider "ibm" {
  region = "us-south"
}

resource "ibm_resource_group" "resource_group" {
  name = "default"
}

resource "ibm_resource_instance" "powervs_service" {
  name              = "Power instance"
  service           = "power-iaas"
  plan              = "power-virtual-server-group"
  location          = "us-south"
  resource_group_id = ibm_resource_group.resource_group.id
}

resource "ibm_pi_volume" "pi_volume_affinity_set" {
  pi_volume_name       = "example-volume"
  pi_cloud_instance_id = ibm_resource_instance.powervs_service.guid
  pi_volume_size       = 100
  pi_affinity_policy   = "affinity"
}

resource "ibm_pi_volume" "pi_volume_volume_set" {
  pi_volume_name       = "example-volume"
  pi_cloud_instance_id = ibm_resource_instance.powervs_service.guid
  pi_volume_size       = 100
  pi_volume_pool       = "volume-pool-name"
}

resource "ibm_pi_volume" "pi_volume_tier1" {
  pi_volume_name       = "example-volume"
  pi_cloud_instance_id = ibm_resource_instance.powervs_service.guid
  pi_volume_size       = 100
  pi_volume_type       = "tier1"
}

resource "ibm_pi_volume" "pi_volume_ssd" {
  pi_volume_name       = "example-volume"
  pi_cloud_instance_id = ibm_resource_instance.powervs_service.guid
  pi_volume_size       = 100
  pi_volume_type       = "ssd"
}

resource "ibm_pi_volume" "pi_volume_tier3" {
  pi_volume_name       = "example-volume"
  pi_cloud_instance_id = ibm_resource_instance.powervs_service.guid
  pi_volume_size       = 100
  pi_volume_type       = "tier3"
}

resource "ibm_pi_volume" "pi_volume_standard" {
  pi_volume_name       = "example-volume"
  pi_cloud_instance_id = ibm_resource_instance.powervs_service.guid
  pi_volume_size       = 100
  pi_volume_type       = "standard"
}

resource "ibm_pi_volume" "pi_volume_tier1_no_usage" {
  pi_volume_name       = "example-volume"
  pi_cloud_instance_id = ibm_resource_instance.powervs_service.guid
  pi_volume_size       = 100
  pi_volume_type       = "tier1"
}
