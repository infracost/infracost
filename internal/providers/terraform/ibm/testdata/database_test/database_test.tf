terraform {
  required_providers {
    ibm = {
      source = "IBM-Cloud/ibm"
    }
  }
}

provider "ibm" {
  region = "us-south"
}

# -------------------------------------------
# POSTGRES
# -------------------------------------------

resource "ibm_database" "postgresql_standard_flavor" {
  name     = "postgres-standard-flavour"
  service  = "databases-for-postgresql"
  plan     = "standard"
  location = "us-south"
  group { # Note: "memory" not allowed when host_flavor is set
    group_id = "member"
    host_flavor {
      id = "m3c.30x240.encrypted"
    }
    disk {
      allocation_mb = 4194304
    }
  }
  configuration = <<CONFIGURATION
  {
    "max_connections": 400
  }
  CONFIGURATION
}

resource "ibm_database" "postgresql_standard" {
  name     = "postgres-standard"
  service  = "databases-for-postgresql"
  plan     = "standard"
  location = "us-south"
  group {
    group_id = "member"
    memory {
      allocation_mb = 114688
    }
    disk {
      allocation_mb = 4194304
    }
    cpu {
      allocation_count = 28
    }
  }
  configuration = <<CONFIGURATION
  {
    "max_connections": 400
  }
  CONFIGURATION
}

# -------------------------------------------
# ELASTICSEARCH
# -------------------------------------------

resource "ibm_database" "elasticsearch_platinum" {
  name     = "elasticsearch-platinum"
  service  = "databases-for-elasticsearch"
  plan     = "platinum"
  location = "us-south"
  group {
    group_id = "member"
    memory {
      allocation_mb = 114688
    }
    disk {
      allocation_mb = 4194304
    }
    cpu {
      allocation_count = 28
    }
  }
}

resource "ibm_database" "elasticsearch_platinum_flavor" {
  name     = "elasticsearch-platinum-flavor"
  service  = "databases-for-elasticsearch"
  plan     = "platinum"
  location = "us-south"
  group { # Note: "memory" not allowed when host_flavor is set
    group_id = "member"
    host_flavor {
      id = "m3c.30x240.encrypted"
    }
    disk {
      allocation_mb = 4194304
    }
  }
}

resource "ibm_database" "elasticsearch_enterprise" {
  name     = "elasticsearch-enterprise"
  service  = "databases-for-elasticsearch"
  plan     = "enterprise"
  location = "us-south"
  group {
    group_id = "member"

    memory {
      allocation_mb = 114688
    }
    disk {
      allocation_mb = 4194304
    }
    cpu {
      allocation_count = 28
    }
  }
}

resource "ibm_database" "elasticsearch_enterprise_flavor" {
  name     = "elasticsearch-enterprise-flavor"
  service  = "databases-for-elasticsearch"
  plan     = "enterprise"
  location = "us-south"
  group { # Note: "memory" not allowed when host_flavor is set
    group_id = "member"
    host_flavor {
      id = "m3c.30x240.encrypted"
    }
    disk {
      allocation_mb = 4194304
    }
  }
}
