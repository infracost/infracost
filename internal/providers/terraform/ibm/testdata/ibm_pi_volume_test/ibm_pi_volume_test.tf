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
  pi_volume_name       = "affinity-volume-set"
  pi_cloud_instance_id = ibm_resource_instance.powervs_service.guid
  pi_volume_size       = 100
  pi_affinity_policy   = "affinity"
  pi_volume_shareable  = false
}

resource "ibm_pi_volume" "pi_volume_volume_set" {
  pi_volume_name       = "volume-pool"
  pi_cloud_instance_id = ibm_resource_instance.powervs_service.guid
  pi_volume_size       = 100
  pi_volume_pool       = "volume-pool-name"
  pi_volume_shareable  = false
}

resource "ibm_pi_volume" "pi_volume_tier0" {
  pi_volume_name       = "tier0-volume"
  pi_cloud_instance_id = ibm_resource_instance.powervs_service.guid
  pi_volume_size       = 100
  pi_volume_type       = "tier0"
  pi_volume_shareable  = false
}

resource "ibm_pi_volume" "pi_volume_tier1" {
  pi_volume_name       = "tier1-volume"
  pi_cloud_instance_id = ibm_resource_instance.powervs_service.guid
  pi_volume_size       = 100
  pi_volume_type       = "tier1"
  pi_volume_shareable  = false
}

resource "ibm_pi_volume" "pi_volume_tier3" {
  pi_volume_name       = "tier3-volume"
  pi_cloud_instance_id = ibm_resource_instance.powervs_service.guid
  pi_volume_size       = 100
  pi_volume_type       = "tier3"
  pi_volume_shareable  = false
}

resource "ibm_pi_volume" "pi_volume_tier5k" {
  pi_volume_name       = "tier5k-volume"
  pi_cloud_instance_id = ibm_resource_instance.powervs_service.guid
  pi_volume_size       = 100
  pi_volume_type       = "tier5k"
  pi_volume_shareable  = false
}

resource "ibm_pi_volume" "pi_volume_tier0_no_usage" {
  pi_volume_name       = "tier0-volume2"
  pi_cloud_instance_id = ibm_resource_instance.powervs_service.guid
  pi_volume_size       = 100
  pi_volume_type       = "tier0"
  pi_volume_shareable  = false
}

resource "ibm_pi_volume" "pi_volume_tier1_no_usage" {
  pi_volume_name       = "tier1-volume2"
  pi_cloud_instance_id = ibm_resource_instance.powervs_service.guid
  pi_volume_size       = 100
  pi_volume_type       = "tier1"
  pi_volume_shareable  = false
}

resource "ibm_pi_volume" "pi_volume_tier3_no_usage" {
  pi_volume_name       = "tier3-volume2"
  pi_cloud_instance_id = ibm_resource_instance.powervs_service.guid
  pi_volume_size       = 100
  pi_volume_type       = "tier3"
  pi_volume_shareable  = false
}

resource "ibm_pi_volume" "pi_volume_tier5k_no_usage" {
  pi_volume_name       = "tier5k-volume2"
  pi_cloud_instance_id = ibm_resource_instance.powervs_service.guid
  pi_volume_size       = 100
  pi_volume_type       = "tier5k"
  pi_volume_shareable  = false
}
