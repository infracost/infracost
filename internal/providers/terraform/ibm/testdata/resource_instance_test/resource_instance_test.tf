terraform {
  required_providers {
    ibm = {
      source  = "IBM-Cloud/ibm"
      version = "1.64.0"
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

resource "ibm_resource_instance" "resource_instance_logdna_lite" {
  name     = "logdna-lite"
  service  = "logdna"
  plan     = "lite"
  location = "us-south"
}

resource "ibm_resource_instance" "resource_instance_logdna_7day" {
  name     = "logdna-7day"
  service  = "logdna"
  plan     = "7-day"
  location = "us-south"
}

resource "ibm_resource_instance" "resource_instance_logdna_7day_no_usage" {
  name     = "logdna-7day-no-usage"
  service  = "logdna"
  plan     = "7-day"
  location = "us-south"
}
resource "ibm_resource_instance" "resource_instance_logdna_14day" {
  name     = "logdna-14day"
  service  = "logdna"
  plan     = "14-day"
  location = "us-south"
}

resource "ibm_resource_instance" "resource_instance_logdna_30day" {
  name     = "logdna-30day"
  service  = "logdna"
  plan     = "30-day"
  location = "us-south"
}

resource "ibm_resource_instance" "resource_instance_logdna_hipaa30day" {
  name     = "logdna-hipaa30day"
  service  = "logdna"
  plan     = "hipaa-30-day"
  location = "us-south"
}

resource "ibm_resource_instance" "resource_instance_activity_tracker_lite" {
  name     = "activity-tracker-lite"
  service  = "logdnaat"
  plan     = "lite"
  location = "us-south"
}

resource "ibm_resource_instance" "resource_instance_activity_tracker_7day" {
  name     = "activity-tracker-7day"
  service  = "logdnaat"
  plan     = "7-day"
  location = "us-south"
}

resource "ibm_resource_instance" "resource_instance_activity_tracker_7day_no_usage" {
  name     = "activity-tracker-7day-no-usage"
  service  = "logdnaat"
  plan     = "7-day"
  location = "us-south"
}

resource "ibm_resource_instance" "resource_instance_monitoring_lite" {
  name     = "sysdig-lite"
  service  = "sysdig-monitor"
  plan     = "lite"
  location = "us-south"
}

resource "ibm_resource_instance" "resource_instance_monitoring_graduated" {
  name     = "sysdig-graduated"
  service  = "sysdig-monitor"
  plan     = "graduated-tier"
  location = "us-south"
}

resource "ibm_resource_instance" "resource_instance_monitoring_graduated_no_usage" {
  name     = "sysdig-graduated-no-usage"
  service  = "sysdig-monitor"
  plan     = "graduated-tier"
  location = "us-south"
}

resource "ibm_resource_instance" "resource_instance_monitoring_graduated_secure" {
  name     = "sysdig-graduated-secure"
  service  = "graduated-tier-sysdig-secure-plus-monitor"
  plan     = "7-day"
  location = "us-south"
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

resource "ibm_resource_instance" "wml_instance_lite" {
  name              = "wml_lite"
  service           = "pm-20"
  plan              = "lite"
  location          = "us-south"
  resource_group_id = "default"
}

resource "ibm_resource_instance" "wml_instance_essentials" {
  name              = "wml_essentials"
  service           = "pm-20"
  plan              = "v2-standard"
  location          = "us-south"
  resource_group_id = "default"
}

resource "ibm_resource_instance" "wml_instance_standard" {
  name              = "wml_standard"
  service           = "pm-20"
  plan              = "v2-professional"
  location          = "us-south"
  resource_group_id = "default"
}

resource "ibm_resource_instance" "wa_instance_lite" {
  name              = "wa_lite"
  service           = "conversation"
  plan              = "lite"
  location          = "us-south"
  resource_group_id = "default"
}

resource "ibm_resource_instance" "wa_instance_trial" {
  name              = "wa_trial"
  service           = "conversation"
  plan              = "plus-trial"
  location          = "us-south"
  resource_group_id = "default"
}

resource "ibm_resource_instance" "wa_instance_plus" {
  name              = "wa_plus"
  service           = "conversation"
  plan              = "plus"
  location          = "us-south"
  resource_group_id = "default"
}

resource "ibm_resource_instance" "wa_instance_enterprise" {
  name              = "wa_enterprise"
  service           = "conversation"
  plan              = "enterprise"
  location          = "us-south"
  resource_group_id = "default"
}

# Watson Discovery
resource "ibm_resource_instance" "watson_discovery_plus" {
  name              = "wd_plus"
  service           = "discovery"
  plan              = "plus"
  location          = "us-south"
  resource_group_id = "default"
}

resource "ibm_resource_instance" "watson_discovery_enterprise" {
  name              = "wd_enterprise"
  service           = "discovery"
  plan              = "enterprise"
  location          = "us-south"
  resource_group_id = "default"
}

# Security and Compliance Center (SCC)
resource "ibm_resource_instance" "scc_standard" {
  name              = "scc_standard"
  service           = "compliance"
  plan              = "security-compliance-center-standard-plan"
  location          = "us-south"
  resource_group_id = "default"
}

resource "ibm_resource_instance" "scc_trial" {
  name              = "scc_trial"
  service           = "compliance"
  plan              = "security-compliance-center-trial-plan"
  location          = "us-south"
  resource_group_id = "default"
}

# Watson Studio
resource "ibm_resource_instance" "watson_studio_professional" {
  name              = "ws_professional"
  service           = "data-science-experience"
  plan              = "professional-v1"
  location          = "us-south"
  resource_group_id = "default"
}

resource "ibm_resource_instance" "watson_studio_lite" {
  name              = "ws_lite"
  service           = "data-science-experience"
  plan              = "free-v1"
  location          = "us-south"
  resource_group_id = "default"
}

# Security and Compliance Center (SCC) Workload Protection
resource "ibm_resource_instance" "sccwp_graduated_tier" {
  name              = "sccwp_graduated_tier"
  service           = "sysdig-secure"
  plan              = "graduated-tier"
  location          = "us-south"
  resource_group_id = "default"
}

# Watsonx.governance
resource "ibm_resource_instance" "watson_governance_lite" {
  name              = "wgov_lite"
  service           = "aiopenscale"
  plan              = "lite"
  location          = "us-south"
  resource_group_id = "default"
}

resource "ibm_resource_instance" "watson_governance_essentials" {
  name              = "wgov_essentials"
  service           = "aiopenscale"
  plan              = "essentials"
  location          = "us-south"
  resource_group_id = "default"
}

resource "ibm_resource_instance" "watson_governance_standard_v2" {
  name              = "wgov_standard_v2"
  service           = "aiopenscale"
  plan              = "standard-v2"
  location          = "us-south"
  resource_group_id = "default"
}

resource "ibm_resource_instance" "dns_svcs_standard" {
  name              = "dns_svcs_standard"
  service           = "dns-svcs"
  plan              = "standard-dns"
  location          = "global"
  resource_group_id = "default"
}

resource "ibm_resource_instance" "messagehub_lite" {
  name              = "messagehub_lite"
  service           = "messagehub"
  plan              = "lite"
  location          = "us-south"
  resource_group_id = "default"
}

resource "ibm_resource_instance" "messagehub_standard" {
  name              = "messagehub_standard"
  service           = "messagehub"
  plan              = "standard"
  location          = "us-south"
  resource_group_id = "default"
}

resource "ibm_resource_instance" "messagehub_enterprise" {
  name              = "messagehub_enterprise"
  service           = "messagehub"
  plan              = "enterprise-3nodes-2tb"
  location          = "us-south"
  resource_group_id = "default"
}

resource "ibm_resource_instance" "messagehub_satellite" {
  name              = "messagehub_satellite"
  service           = "messagehub"
  plan              = "satellite"
  location          = "satcon_dal"
  resource_group_id = "default"
}
