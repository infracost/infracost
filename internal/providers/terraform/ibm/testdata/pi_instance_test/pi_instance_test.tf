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
  zone   = "dal12"
}

locals {
  service_type = "power-iaas"
}

resource "ibm_resource_group" "resource_group" {
  name = "default"
}

resource "ibm_resource_instance" "powervs_service" {
  name              = "Power instance"
  service           = local.service_type
  plan              = "power-virtual-server-group"
  location          = "us-south"
  resource_group_id = ibm_resource_group.resource_group.id
}


resource "ibm_pi_image" "aix_image" {
  pi_image_name             = "7200-05-03"
  pi_cloud_instance_id      = ibm_resource_instance.powervs_service.guid
  pi_image_bucket_name      = "images-public-bucket"
  pi_image_bucket_access    = "public"
  pi_image_bucket_region    = "us-south"
  pi_image_bucket_file_name = "rhcos-48-07222021.ova.gz"
  pi_image_storage_type     = "tier1"
}

resource "ibm_pi_image" "ibmi_image" {
  pi_image_name             = "IBMi-71-11-2924-4"
  pi_cloud_instance_id      = ibm_resource_instance.powervs_service.guid
  pi_image_bucket_name      = "images-public-bucket"
  pi_image_bucket_access    = "public"
  pi_image_bucket_region    = "us-south"
  pi_image_bucket_file_name = "rhcos-48-07222021.ova.gz"
  pi_image_storage_type     = "tier1"
}

resource "ibm_pi_image" "hana_image" {
  pi_image_name             = "SLES15-SP2-SAP"
  pi_cloud_instance_id      = ibm_resource_instance.powervs_service.guid
  pi_image_bucket_name      = "images-public-bucket"
  pi_image_bucket_access    = "public"
  pi_image_bucket_region    = "us-south"
  pi_image_bucket_file_name = "rhcos-48-07222021.ova.gz"
  pi_image_storage_type     = "tier1"
}

resource "ibm_pi_image" "netweaver_image" {
  pi_image_name             = "SLES15-SP2-SAP-NETWEAVER"
  pi_cloud_instance_id      = ibm_resource_instance.powervs_service.guid
  pi_image_bucket_name      = "images-public-bucket"
  pi_image_bucket_access    = "public"
  pi_image_bucket_region    = "us-south"
  pi_image_bucket_file_name = "rhcos-48-07222021.ova.gz"
  pi_image_storage_type     = "tier1"
}

resource "ibm_pi_key" "key" {
  pi_key_name          = "testkey"
  pi_ssh_key           = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCKVmnMOlHKcZK8tpt3MP1lqOLAcqcJzhsvJcjscgVERRN7/9484SOBJ3HSKxxNG5JN8owAjy5f9yYwcUg+JaUVuytn5Pv3aeYROHGGg+5G346xaq3DAwX6Y5ykr2fvjObgncQBnuU5KHWCECO/4h8uWuwh/kfniXPVjFToc+gnkqA+3RKpAecZhFXwfalQ9mMuYGFxn+fwn8cYEApsJbsEmb0iJwPiZ5hjFC8wREuiTlhPHDgkBLOiycd20op2nXzDbHfCHInquEe/gYxEitALONxm0swBOwJZwlTDOB7C6y2dzlrtxr1L59m7pCkWI4EtTRLvleehBoj3u7jB4usR"
  pi_cloud_instance_id = ibm_resource_instance.powervs_service.guid
}

resource "ibm_pi_instance" "aix-shared-s922-instance" {
  pi_memory            = "1"
  pi_processors        = "1"
  pi_instance_name     = "aix-shared-s922"
  pi_proc_type         = "shared"
  pi_image_id          = ibm_pi_image.aix_image.id
  pi_key_pair_name     = ibm_pi_key.key.key_id
  pi_sys_type          = "s922"
  pi_cloud_instance_id = ibm_resource_instance.powervs_service.guid
  pi_pin_policy        = "none"
  pi_health_status     = "WARNING"
  pi_storage_type      = ibm_pi_image.aix_image.pi_image_storage_type
  pi_network {
    network_id = "test-id"
  }
}

resource "ibm_pi_instance" "ibmi-dedicated-e980-instance" {
  pi_memory            = "1"
  pi_processors        = "1"
  pi_instance_name     = "ibmi-dedicated-e980"
  pi_proc_type         = "dedicated"
  pi_image_id          = ibm_pi_image.ibmi_image.id
  pi_key_pair_name     = ibm_pi_key.key.key_id
  pi_sys_type          = "e980"
  pi_cloud_instance_id = ibm_resource_instance.powervs_service.guid
  pi_pin_policy        = "none"
  pi_health_status     = "WARNING"
  pi_storage_type      = ibm_pi_image.ibmi_image.pi_image_storage_type
  pi_network {
    network_id = "test-id"
  }
}

resource "ibm_pi_instance" "hana-dedicated-e980-instance" {
  pi_instance_name     = "hana-dedicated-e980"
  pi_image_id          = ibm_pi_image.hana_image.id
  pi_key_pair_name     = ibm_pi_key.key.key_id
  pi_sys_type          = "e980"
  pi_cloud_instance_id = ibm_resource_instance.powervs_service.guid
  pi_pin_policy        = "none"
  pi_health_status     = "WARNING"
  pi_storage_type      = ibm_pi_image.ibmi_image.pi_image_storage_type
  pi_sap_profile_id    = "ush1-4x128"
  pi_network {
    network_id = "test-id"
  }
}

resource "ibm_pi_instance" "netweaver-shared-s922-instance" {
  pi_memory            = "1"
  pi_processors        = "1"
  pi_instance_name     = "netweaver-shared-s922"
  pi_proc_type         = "shared"
  pi_image_id          = ibm_pi_image.netweaver_image.id
  pi_key_pair_name     = ibm_pi_key.key.key_id
  pi_sys_type          = "s922"
  pi_cloud_instance_id = ibm_resource_instance.powervs_service.guid
  pi_pin_policy        = "none"
  pi_health_status     = "WARNING"
  pi_storage_type      = ibm_pi_image.ibmi_image.pi_image_storage_type
  pi_network {
    network_id = "test-id"
  }
}

resource "ibm_pi_instance" "netweaver-shared-s922-no-usage-specified-instance" {
  pi_memory            = "1"
  pi_processors        = "1"
  pi_instance_name     = "netweaver-shared-s922"
  pi_proc_type         = "shared"
  pi_image_id          = ibm_pi_image.netweaver_image.id
  pi_key_pair_name     = ibm_pi_key.key.key_id
  pi_sys_type          = "s922"
  pi_cloud_instance_id = ibm_resource_instance.powervs_service.guid
  pi_pin_policy        = "none"
  pi_health_status     = "WARNING"
  pi_storage_type      = ibm_pi_image.ibmi_image.pi_image_storage_type
  pi_network {
    network_id = "test-id"
  }
}

