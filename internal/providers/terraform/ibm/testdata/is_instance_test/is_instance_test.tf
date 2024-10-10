
terraform {
  required_providers {
    ibm = {
      source = "IBM-Cloud/ibm"
    }
    random = {
      source = "hashicorp/random"
    }
  }
}

provider "ibm" {
  region = "us-south"
}

# Access random string generated with random_string.unique_identifier.result
resource "random_string" "unique_identifier" {
  length  = 6
  special = false
  upper   = false
}

resource "ibm_resource_group" "resource_group" {
  name = "rg-${random_string.unique_identifier.result}"
}

resource "ibm_is_vpc" "vpc" {
  name           = "vpc-${random_string.unique_identifier.result}"
  resource_group = ibm_resource_group.resource_group.id
}

resource "ibm_is_subnet" "subnet" {
  name            = "subnet-${random_string.unique_identifier.result}"
  ipv4_cidr_block = "10.240.0.0/24"
  resource_group  = ibm_resource_group.resource_group.id
  vpc             = ibm_is_vpc.vpc.id
  zone            = "us-south-1"
}

resource "ibm_is_ssh_key" "ssh_key" {
  name           = "ssh-key-${random_string.unique_identifier.result}"
  public_key     = file("~/.ssh/id_ed25519.pub")
  resource_group = ibm_resource_group.resource_group.id
  type           = "ed25519"
}

resource "ibm_is_instance" "vsi" {
  for_each = toset(local.profiles)
  name     = "vsi-instance-${random_string.unique_identifier.result}-${each.key}"
  # name    = "vsi-instance-${random_string.unique_identifier.result}"
  image = "r006-f137ea64-0d27-4d81-afe0-353fd0557e81"
  keys  = [ibm_is_ssh_key.ssh_key.id]
  # profile = "cx2-2x4"
  profile        = each.key
  resource_group = ibm_resource_group.resource_group.id
  vpc            = ibm_is_vpc.vpc.id
  zone           = "us-south-1"
  primary_network_interface {
    subnet = ibm_is_subnet.subnet.id
  }
  network_interfaces {
    name   = "eth1"
    subnet = ibm_is_subnet.subnet.id
  }
}

resource "ibm_is_instance" "vsi_boot_volume" {
  for_each = toset(local.profiles)
  name     = "vsi-instance-boot-volume-${random_string.unique_identifier.result}-${each.key}"
  # name    = "vsi-instance-boot-volume-${random_string.unique_identifier.result}"
  image = "r006-f137ea64-0d27-4d81-afe0-353fd0557e81"
  keys  = [ibm_is_ssh_key.ssh_key.id]
  # profile = "cx2-2x4"
  profile        = each.key
  resource_group = ibm_resource_group.resource_group.id
  vpc            = ibm_is_vpc.vpc.id
  zone           = "us-south-1"
  primary_network_interface {
    subnet = ibm_is_subnet.subnet.id
  }
  network_interfaces {
    name   = "eth1"
    subnet = ibm_is_subnet.subnet.id
  }
  boot_volume {
    name = "boot-volume-label"
    size = 250
  }
}

resource "ibm_is_instance" "vsi_dedicated_host" {
  for_each = toset(local.profiles)
  name     = "vsi-instance-dedicated-host-${random_string.unique_identifier.result}-${each.key}"
  # name    = "vsi-instance-dedicated-host-${random_string.unique_identifier.result}"
  image = "r006-f137ea64-0d27-4d81-afe0-353fd0557e81"
  keys  = [ibm_is_ssh_key.ssh_key.id]
  # profile = "cx2-2x4"
  profile        = each.key
  resource_group = ibm_resource_group.resource_group.id
  vpc            = ibm_is_vpc.vpc.id
  zone           = "us-south-1"
  primary_network_interface {
    subnet = ibm_is_subnet.subnet.id
  }
  network_interfaces {
    name   = "eth1"
    subnet = ibm_is_subnet.subnet.id
  }
  dedicated_host = ibm_is_dedicated_host.dedicated_host.id
}

resource "ibm_is_dedicated_host" "dedicated_host" {
  profile        = "bx2d-host-152x608"
  name           = "example-dedicated-host-01"
  host_group     = ibm_is_dedicated_host_group.dedicated_host_group.id
  resource_group = ibm_resource_group.resource_group.id
}

resource "ibm_is_dedicated_host_group" "dedicated_host_group" {
  family         = "compute"
  class          = "cx2"
  zone           = "us-south-1"
  name           = "example-dh-group-01"
  resource_group = ibm_resource_group.resource_group.id
}

locals {
  profiles = [
    "cx2-2x4",
    "cx2d-2x4",
    "cx3d-2x5",
    "cx3dc-2x5",
    "bx2-2x8",
    "bx2a-2x8",
    "bx2d-2x8",
    "bx3d-2x10",
    "bx3dc-2x10",
    "mx2-2x16",
    "mx2d-2x16",
    "ox2-2x16",
    "mx3d-2x20",
    "vx2d-2x28",
    "ux2d-2x56",
    "cx2-4x8",
    "cx2d-4x8",
    "cx3d-4x10",
    "cx3dc-4x10",
    "bx2-4x16",
    "bx2a-4x16",
    "bx2d-4x16",
    "bx3d-4x20",
    "bx3dc-4x20",
    "mx2-4x32",
    "mx2d-4x32",
    "ox2-4x32",
    "mx3d-4x40",
    "vx2d-4x56",
    "ux2d-4x112",
    "cx2-8x16",
    "cx2d-8x16",
    "cx3d-8x20",
    "cx3dc-8x20",
    "bx2-8x32",
    "bx2a-8x32",
    "bx2d-8x32",
    "bx3d-8x40",
    "bx3dc-8x40",
    "mx2-8x64",
    "mx2d-8x64",
    "ox2-8x64",
    "gx2-8x64x1v100",
    "mx3d-8x80",
    "vx2d-8x112",
    "ux2d-8x224",
    "cx2-16x32",
    "cx2d-16x32",
    "cx3d-16x40",
    "cx3dc-16x40",
    "bx2-16x64",
    "bx2a-16x64",
    "bx2d-16x64",
    "bx3d-16x80",
    "gx3-16x80x1l4",
    "bx3dc-16x80",
    "mx2-16x128",
    "mx2d-16x128",
    "ox2-16x128",
    "gx2-16x128x1v100",
    "gx2-16x128x2v100",
    "mx3d-16x160",
    "vx2d-16x224",
    "ux2d-16x448",
    "cx3d-24x60",
    "cx3dc-24x60",
    "bx3d-24x120",
    "bx3dc-24x120",
    "gx3-24x120x1l40s",
    "mx3d-24x240",
    "cx2-32x64",
    "cx2d-32x64",
    "cx3d-32x80",
    "cx3dc-32x80",
    "bx2-32x128",
    "bx2a-32x128",
    "bx2d-32x128",
    "bx3d-32x160",
    "gx3-32x160x2l4",
    "bx3dc-32x160",
    "mx2-32x256",
    "mx2d-32x256",
    "ox2-32x256",
    "gx2-32x256x2v100",
    "mx3d-32x320",
    "ux2d-36x1008",
    "vx2d-44x616",
    "cx2-48x96",
    "cx2d-48x96",
    "cx3d-48x120",
    "cx3dc-48x120",
    "bx2-48x192",
    "bx2a-48x192",
    "bx2d-48x192",
    "bx3d-48x240",
    "bx3dc-48x240",
    "gx3-48x240x2l40s",
    "mx2-48x384",
    "mx2d-48x384",
    "mx3d-48x480",
    "ux2d-48x1344",
    "cx2-64x128",
    "cx2d-64x128",
    "cx3d-64x160",
    "cx3dc-64x160",
    "bx2-64x256",
    "bx2d-64x256",
    "bx3d-64x320",
    "gx3-64x320x4l4",
    "bx3dc-64x320",
    "mx2-64x512",
    "mx2d-64x512",
    "ox2-64x512",
    "mx3d-64x640",
    "ux2d-72x2016",
    "vx2d-88x1232",
    "cx2-96x192",
    "cx2d-96x192",
    "cx3d-96x240",
    "cx3dc-96x240",
    "bx2-96x384",
    "bx2a-96x384",
    "bx2d-96x384",
    "bx3d-96x480",
    "bx3dc-96x480",
    "mx2-96x768",
    "mx2d-96x768",
    "ox2-96x768",
    "mx3d-96x960",
    "ux2d-100x2800",
    "cx2-128x256",
    "cx2d-128x256",
    "cx3d-128x320",
    "cx3dc-128x320",
    "bx2-128x512",
    "bx2a-128x512",
    "bx2d-128x512",
    "bx3d-128x640",
    "mx2-128x1024",
    "mx2d-128x1024",
    "ox2-128x1024",
    "mx3d-128x1280",
    "vx2d-144x2016",
    "gx3d-160x1792x8h100",
    "cx3d-176x440",
    "bx3d-176x880",
    "mx3d-176x1760",
    "vx2d-176x2464",
    "ux2d-200x5600",
    "bx2a-228x912"
  ]
}
