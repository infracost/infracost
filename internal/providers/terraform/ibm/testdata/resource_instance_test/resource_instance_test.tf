terraform {
  required_providers {
    ibm = {
      source = "IBM-Cloud/ibm"
      version = "~> 1.40.0"
    }
  }
}

provider "ibm" {
    region = "us-south"
}

resource "ibm_resource_instance" "resource_instance_kms" {
  name              = "test"
  service           = "kms"
  plan              = "tiered-pricing"
  location          = "us-south"
  resource_group_id = "default"
}

resource "ibm_resource_instance" "resource_instance_secrets_manager" {
  name              = "test"
  service           = "secrets-manager"
  plan              = "standard"
  location          = "us-south"
  resource_group_id = "default"
}

resource "ibm_resource_instance" "resource_instance_appid" {
  name              = "test"
  service           = "appid"
  plan              = "graduated-tier"
  location          = "us-south"
  resource_group_id = "default"
}

resource "ibm_resource_instance" "resource_instance_power_iaas" {
  name              = "test"
  service           = "power-iaas"
  plan              = "power-virtual-server-group"
  location          = "us-south"
  resource_group_id = "default"
}

resource "ibm_resource_instance" "resource_instance_appconnect_ent" {
  name              = "appconnect-ent"
  service           = "appconnect"
  plan              = "appconnectplanenterprise"
  location          = "us-south"
}

resource "ibm_resource_instance" "resource_instance_appconnect_pro" {
  name              = "appconnect-pro"
  service           = "appconnect"
  plan              = "appconnectplanprofessional"
  location          = "us-south"
}

resource "ibm_resource_instance" "resource_instance_appconnect_lite" {
  name              = "appconnect-lite"
  service           = "appconnect"
  plan              = "lite"
  location          = "us-south"
}
