
terraform {
  required_providers {
    ibm = {
      source  = "IBM-Cloud/ibm"
      version = "1.63.0"
    }
  }
}

provider "ibm" {
  region = "us-east"
}

resource "ibm_is_vpc" "testVpc" {
  name = "test-vpc"
}

resource "ibm_is_subnet" "testSubnet" {
  name            = "test-subnet"
  vpc             = ibm_is_vpc.testVpc.id
  zone            = "us-east-1"
  ipv4_cidr_block = "10.240.0.0/24"
}

resource "ibm_is_ssh_key" "testSshKey" {
  name       = "test-ssh"
  public_key = file("~/.ssh/id_rsa.pub")
}

resource "ibm_is_instance" "testBalancedInstance" {
  name    = "test-instance-1"
  image   = "r006-f137ea64-0d27-4d81-afe0-353fd0557e81"
  profile = "bx2d-32x128"

  primary_network_interface {
    subnet = ibm_is_subnet.testSubnet.id
  }

  network_interfaces {
    name   = "eth1"
    subnet = ibm_is_subnet.testSubnet.id
  }

  vpc  = ibm_is_vpc.testVpc.id
  zone = "us-east-1"
  keys = [ibm_is_ssh_key.testSshKey.id]
}

resource "ibm_is_instance" "testComputeInstance" {
  name    = "test-instance-2"
  image   = "r006-f137ea64-0d27-4d81-afe0-353fd0557e81"
  profile = "cx2-2x4"

  primary_network_interface {
    subnet = ibm_is_subnet.testSubnet.id
  }

  network_interfaces {
    name   = "eth1"
    subnet = ibm_is_subnet.testSubnet.id
  }

  vpc  = ibm_is_vpc.testVpc.id
  zone = "us-east-1"
  keys = [ibm_is_ssh_key.testSshKey.id]
}

resource "ibm_is_instance" "testGpuInstance" {
  name    = "test-instance-3"
  image   = "r006-f137ea64-0d27-4d81-afe0-353fd0557e81"
  profile = "gx2-16x128x2v100"

  primary_network_interface {
    subnet = ibm_is_subnet.testSubnet.id
  }

  network_interfaces {
    name   = "eth1"
    subnet = ibm_is_subnet.testSubnet.id
  }

  vpc  = ibm_is_vpc.testVpc.id
  zone = "us-east-1"
  keys = [ibm_is_ssh_key.testSshKey.id]
}

resource "ibm_is_instance" "testIbmZInstance" {
  name    = "test-instance-4"
  image   = "r006-f137ea64-0d27-4d81-afe0-353fd0557e81"
  profile = "bz2-16x64"

  primary_network_interface {
    subnet = ibm_is_subnet.testSubnet.id
  }

  network_interfaces {
    name   = "eth1"
    subnet = ibm_is_subnet.testSubnet.id
  }

  vpc  = ibm_is_vpc.testVpc.id
  zone = "us-east-1"
  keys = [ibm_is_ssh_key.testSshKey.id]
}

resource "ibm_is_instance" "testInstanceWithoutUsage" {
  name    = "test-instance-5"
  image   = "r006-f137ea64-0d27-4d81-afe0-353fd0557e81"
  profile = "cx2-2x4"

  primary_network_interface {
    subnet = ibm_is_subnet.testSubnet.id
  }

  network_interfaces {
    name   = "eth1"
    subnet = ibm_is_subnet.testSubnet.id
  }

  vpc  = ibm_is_vpc.testVpc.id
  zone = "us-east-1"
  keys = [ibm_is_ssh_key.testSshKey.id]
}


resource "ibm_is_instance" "testBalancedInstanceWithBootVolume" {
  name    = "test-instance-6"
  image   = "r006-f137ea64-0d27-4d81-afe0-353fd0557e81"
  profile = "bx2-8x32"

  primary_network_interface {
    subnet = ibm_is_subnet.testSubnet.id
  }

  network_interfaces {
    name   = "eth1"
    subnet = ibm_is_subnet.testSubnet.id
  }

  vpc  = ibm_is_vpc.testVpc.id
  zone = "us-east-1"
  keys = [ibm_is_ssh_key.testSshKey.id]

  boot_volume {
    name = "boot-volume-label"
    size = 150
  }
}
