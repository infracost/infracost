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

resource "ibm_is_image" "testImage" {
  name             = "test-image"
  href             = "cos://us-south/buckettesttest/livecd.ubuntu-cpc.azure.vhd"
  operating_system = "ubuntu-16-04-amd64"
}

resource "ibm_is_subnet" "testIsSubnet" {
  name = "test-is-subnet"
  vpc  = ibm_is_vpc.testVpc.id
  zone = "us-south-1"
}

resource "ibm_is_ssh_key" "testIsShhKey" {
  name       = "test-is-ssh-key"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCKVmnMOlHKcZK8tpt3MP1lqOLAcqcJzhsvJcjscgVERRN7/9484SOBJ3HSKxxNG5JN8owAjy5f9yYwcUg+JaUVuytn5Pv3aeYROHGGg+5G346xaq3DAwX6Y5ykr2fvjObgncQBnuU5KHWCECO/4h8uWuwh/kfniXPVjFToc+gnkqA+3RKpAecZhFXwfalQ9mMuYGFxn+fwn8cYEApsJbsEmb0iJwPiZ5hjFC8wREuiTlhPHDgkBLOiycd20op2nXzDbHfCHInquEe/gYxEitALONxm0swBOwJZwlTDOB7C6y2dzlrtxr1L59m7pCkWI4EtTRLvleehBoj3u7jB4usR"
}

resource "ibm_is_instance" "testInstance" {
  name    = "test-is-instance"
  image   = ibm_is_image.testImage.id
  profile = "bx2-2x8"

  primary_network_interface {
    subnet = ibm_is_subnet.testIsSubnet.id
  }

  vpc  = ibm_is_vpc.testVpc.id
  zone = "us-south-1"
  keys = [ibm_is_ssh_key.testIsShhKey.id]
}

resource "ibm_is_floating_ip" "testFloatingIp" {
  name   = "test-is-floating-ip"
  target = ibm_is_instance.testInstance.primary_network_interface[0].id
}
