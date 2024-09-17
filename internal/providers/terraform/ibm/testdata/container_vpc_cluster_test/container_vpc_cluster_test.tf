terraform {
  required_providers {
    ibm = {
      source  = "IBM-Cloud/ibm"
      version = "1.63.0"
    }
  }
}
provider "ibm" {
  generation       = 2
  region           = "eu-de"
  ibmcloud_timeout = "1"
  max_retries      = "1"
}

resource "ibm_is_vpc" "vpc1" {
  name = "myvpc"
}

resource "ibm_is_subnet" "subnet1" {
  name                     = "mysubnet1"
  vpc                      = ibm_is_vpc.vpc1.id
  zone                     = "eu-de-1"
  total_ipv4_address_count = 256
}

resource "ibm_is_subnet" "subnet2" {
  name                     = "mysubnet2"
  vpc                      = ibm_is_vpc.vpc1.id
  zone                     = "eu-de-2"
  total_ipv4_address_count = 256
}

resource "ibm_container_vpc_cluster" "cluster" {
  name         = "mycluster"
  vpc_id       = ibm_is_vpc.vpc1.id
  flavor       = "bx2.4x16"
  worker_count = 3
  kube_version = "1.17.5"
  zones {
    subnet_id = ibm_is_subnet.subnet1.id
    name      = "eu-de-1"
  }
  zones {
    subnet_id = ibm_is_subnet.subnet1.id
    name      = "eu-de-2"
  }
}

resource "ibm_container_vpc_cluster" "cluster_without_usage" {
  name         = "mycluster-without-usage"
  vpc_id       = ibm_is_vpc.vpc1.id
  flavor       = "bx2.4x16"
  worker_count = 3
  kube_version = "1.17.5"
  zones {
    subnet_id = ibm_is_subnet.subnet1.id
    name      = "eu-de-1"
  }
  zones {
    subnet_id = ibm_is_subnet.subnet1.id
    name      = "eu-de-2"
  }
}

resource "ibm_container_vpc_cluster" "roks_cluster_with_usage" {
  name         = "myrokscluster-with-usage"
  vpc_id       = ibm_is_vpc.vpc1.id
  flavor       = "bx2.4x16"
  worker_count = 3
  kube_version = "4.13_openshift"
  zones {
    subnet_id = ibm_is_subnet.subnet1.id
    name      = "eu-de-1"
  }
  zones {
    subnet_id = ibm_is_subnet.subnet1.id
    name      = "eu-de-2"
  }
}
