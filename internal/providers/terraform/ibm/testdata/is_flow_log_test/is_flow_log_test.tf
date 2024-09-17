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

resource "ibm_is_vpc" "testVpc" {
  name = "test-vpc"
}

resource "ibm_is_image" "testIsImage" {
  name             = "test-image"
  href             = "cos://us-south/buckettesttest/livecd.ubuntu-cpc.azure.vhd"
  operating_system = "ubuntu-16-04-amd64"
}

resource "ibm_is_subnet" "testSubnet" {
  name = "test-subnet"
  vpc  = ibm_is_vpc.testVpc.id
  zone = "us-south-1"
}

resource "ibm_is_ssh_key" "testSshKey" {
  name       = "test-ssh"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCKVmnMOlHKcZK8tpt3MP1lqOLAcqcJzhsvJcjscgVERRN7/9484SOBJ3HSKxxNG5JN8owAjy5f9yYwcUg+JaUVuytn5Pv3aeYROHGGg+5G346xaq3DAwX6Y5ykr2fvjObgncQBnuU5KHWCECO/4h8uWuwh/kfniXPVjFToc+gnkqA+3RKpAecZhFXwfalQ9mMuYGFxn+fwn8cYEApsJbsEmb0iJwPiZ5hjFC8wREuiTlhPHDgkBLOiycd20op2nXzDbHfCHInquEe/gYxEitALONxm0swBOwJZwlTDOB7C6y2dzlrtxr1L59m7pCkWI4EtTRLvleehBoj3u7jB4usR"
}


resource "ibm_is_instance" "testIsInstance" {
  name    = "test-is-instance"
  image   = ibm_is_image.testIsImage.id
  profile = "bx2-2x8"

  primary_network_interface {
    subnet = ibm_is_subnet.testSubnet.id
  }

  vpc  = ibm_is_vpc.testVpc.id
  zone = "us-south-1"
  keys = [ibm_is_ssh_key.testSshKey.id]
}


resource "ibm_resource_instance" "testResourceInstance" {
  name     = "test-cos-instance"
  service  = "cloud-object-storage"
  plan     = "standard"
  location = "global"
}

resource "ibm_cos_bucket" "testCosBucket" {
  bucket_name          = "us-south-bucket-vpc1"
  resource_instance_id = ibm_resource_instance.testResourceInstance.id
  region_location      = "us-south"
  storage_class        = "standard"
}

resource "ibm_is_flow_log" "testFlowLog" {
  depends_on     = [ibm_cos_bucket.testCosBucket]
  name           = "test-instance-flow-log"
  target         = ibm_is_instance.testIsInstance.id
  active         = true
  storage_bucket = ibm_cos_bucket.testCosBucket.bucket_name
}

resource "ibm_is_flow_log" "testFlowLogWithoutUsage" {
  depends_on     = [ibm_cos_bucket.testCosBucket]
  name           = "test-instance-flow-log"
  target         = ibm_is_instance.testIsInstance.id
  active         = true
  storage_bucket = ibm_cos_bucket.testCosBucket.bucket_name
}
