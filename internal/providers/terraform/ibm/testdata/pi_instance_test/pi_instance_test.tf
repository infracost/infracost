terraform {
  required_providers {
    ibm = {
      source  = "IBM-Cloud/ibm"
      version = "~> 1.40.0"
    }
  }
}

provider "ibm" {
  region           = "us-south"
  zone             = "dal12"
  ibmcloud_api_key = "AXrURglaTG9MApQXfqmSoME0dkcBAyv1v9Hw5vljab8y"
}

resource "ibm_pi_image" "powerimages" {
  pi_image_name        = "7200-05-03"
  pi_cloud_instance_id = "51e1879c-bcbe-4ee1-a008-49cdba0eaf60"
  pi_image_bucket_name = "images-public-bucket"
  pi_image_bucket_access = "public"
  pi_image_bucket_region = "us-south"
  pi_image_bucket_file_name = "rhcos-48-07222021.ova.gz"
  pi_image_storage_type = "tier1"
}
resource "ibm_pi_key" "key" {
  pi_key_name          = "testkey"
  pi_ssh_key           = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCKVmnMOlHKcZK8tpt3MP1lqOLAcqcJzhsvJcjscgVERRN7/9484SOBJ3HSKxxNG5JN8owAjy5f9yYwcUg+JaUVuytn5Pv3aeYROHGGg+5G346xaq3DAwX6Y5ykr2fvjObgncQBnuU5KHWCECO/4h8uWuwh/kfniXPVjFToc+gnkqA+3RKpAecZhFXwfalQ9mMuYGFxn+fwn8cYEApsJbsEmb0iJwPiZ5hjFC8wREuiTlhPHDgkBLOiycd20op2nXzDbHfCHInquEe/gYxEitALONxm0swBOwJZwlTDOB7C6y2dzlrtxr1L59m7pCkWI4EtTRLvleehBoj3u7jB4usR"
  pi_cloud_instance_id = "51e1879c-bcbe-4ee1-a008-49cdba0eaf60"
}

resource "ibm_pi_instance" "test-instance" {
  pi_memory            = "4"
  pi_processors        = "0.25"
  pi_instance_name     = "test-vm"
  pi_proc_type         = "shared"
  pi_image_id          = ibm_pi_image.powerimages.id
  pi_key_pair_name     = ibm_pi_key.key.key_id
  pi_sys_type          = "s922"
  pi_cloud_instance_id = "51e1879c-bcbe-4ee1-a008-49cdba0eaf60"
  pi_pin_policy        = "none"
  pi_health_status     = "WARNING"
  pi_storage_type  = ibm_pi_image.powerimages.pi_image_storage_type
  pi_network {
    network_id = "test-id"
  }
}
