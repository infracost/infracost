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

resource "ibm_is_share" "testNFS" {
  name    = "test-share"
  size    = 100
  profile = "dp2"
  zone    = "us-south-2"
  iops    = 100
}

resource "ibm_is_share" "testReplica" {
  zone                  = "us-south-3"
  source_share          = ibm_is_share.testNFS.id
  name                  = "test-replica-share"
  profile               = "dp2"
  replication_cron_spec = "0 */5 * * *"
}

resource "ibm_is_share_mount_target" "testNFSMount" {
  share = ibm_is_share.testNFS.id
  vpc   = ibm_is_vpc.testVpc.id
  name  = "test-share-target"
}

resource "ibm_is_share" "testNFS2withInlineReplica" {
  name    = "test-share-2"
  size    = 200
  profile = "dp2"
  zone    = "us-south-2"

  replica_share {
    name                  = "test-replica-share-2"
    replication_cron_spec = "0 */5 * * *"
    profile               = "dp2"
    zone                  = "us-south-3"
  }
}
