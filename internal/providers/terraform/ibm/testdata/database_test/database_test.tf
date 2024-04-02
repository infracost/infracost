
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

resource "ibm_database" "test_db1" {
  name     = "demo-postgres"
  service  = "databases-for-postgresql"
  plan     = "standard"
  location = "eu-gb"

  group {
    group_id = "member"
    memory {
      allocation_mb = 12288
    }
    disk {
      allocation_mb = 131072
    }
    cpu {
      allocation_count = 3
    }
  }
  configuration = <<CONFIGURATION
  {
    "max_connections": 400
  }
  CONFIGURATION
}

resource "ibm_database" "test_db2" {
  name     = "demo-postgres2"
  service  = "databases-for-postgresql"
  plan     = "standard"
  location = "eu-gb"

  group {
    group_id = "member"
    memory {
      allocation_mb = 15360
    }
    members {
      allocation_count = 4
    }
  }
  configuration = <<CONFIGURATION
  {
    "max_connections": 400
  }
  CONFIGURATION
}
