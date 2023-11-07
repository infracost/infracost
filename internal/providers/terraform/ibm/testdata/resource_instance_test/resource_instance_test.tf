terraform {
  required_providers {
    ibm = {
      source = "IBM-Cloud/ibm"
      version = "1.58.0"
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

resource "ibm_resource_instance" "resource_instance_logdna_lite" {
  name              = "logdna-lite"
  service           = "logdna"
  plan              = "lite"
  location          = "us-south"
}

resource "ibm_resource_instance" "resource_instance_logdna_7day" {
  name              = "logdna-7day"
  service           = "logdna"
  plan              = "7-day"
  location          = "us-south"
}

resource "ibm_resource_instance" "resource_instance_logdna_7day_no_usage" {
  name              = "logdna-7day-no-usage"
  service           = "logdna"
  plan              = "7-day"
  location          = "us-south"
}

resource "ibm_resource_instance" "resource_instance_activity_tracker_lite" {
  name              = "activity-tracker-lite"
  service           = "logdnaat"
  plan              = "lite"
  location          = "us-south"
}

resource "ibm_resource_instance" "resource_instance_activity_tracker_7day" {
  name              = "activity-tracker-7day"
  service           = "logdnaat"
  plan              = "7-day"
  location          = "us-south"
}

resource "ibm_resource_instance" "resource_instance_activity_tracker_7day_no_usage" {
  name              = "activity-tracker-7day-no-usage"
  service           = "logdnaat"
  plan              = "7-day"
  location          = "us-south"
}

resource "ibm_resource_instance" "resource_instance_monitoring_lite" {
  name              = "sysdig-lite"
  service           = "sysdig-monitor"
  plan              = "lite"
  location          = "us-south"
}

resource "ibm_resource_instance" "resource_instance_monitoring_graduated" {
  name              = "sysdig-graduated"
  service           = "sysdig-monitor"
  plan              = "graduated-tier"
  location          = "us-south"
}

resource "ibm_resource_instance" "resource_instance_monitoring_graduated_no_usage" {
  name              = "sysdig-graduated-no-usage"
  service           = "sysdig-monitor"
  plan              = "graduated-tier"
  location          = "us-south"
}

resource "ibm_resource_instance" "resource_instance_monitoring_graduated_secure" {
  name              = "sysdig-graduated-secure"
  service           = "graduated-tier-sysdig-secure-plus-monitor"
  plan              = "7-day"
  location          = "us-south"
}

resource "ibm_resource_instance" "cd_instance_professional" {
  name              = "cd_professional"
  service           = "continuous-delivery"
  plan              = "professional"
  location          = "us-south"
  resource_group_id = "default"
}

resource "ibm_resource_instance" "cd_instance_lite" {
  name              = "cd_lite"
  service           = "continuous-delivery"
  plan              = "lite"
  location          = "us-south"
  resource_group_id = "default"
}