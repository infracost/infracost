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

resource "ibm_database" "test_es_enterprise_db1" {
  name     = "demo-es-enterprise"
  service  = "databases-for-elasticsearch"
  plan     = "enterprise"
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
}

resource "ibm_database" "test_es_enterprise_db2" {
  name     = "demo-es-enterprise2"
  service  = "databases-for-elasticsearch"
  plan     = "enterprise"
  location = "eu-gb"

  group {
    group_id = "member"
    members {
      allocation_count = 4
    }
    memory {
      allocation_mb = 15360
    }
  }
}

resource "ibm_database" "test_es_platinum_db1" {
  name     = "demo-es-platinum"
  service  = "databases-for-elasticsearch"
  plan     = "platinum"
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
}

resource "ibm_database" "test_es_platinum_db2" {
  name     = "demo-es-platinum2"
  service  = "databases-for-elasticsearch"
  plan     = "platinum"
  location = "eu-gb"

  group {
    group_id = "member"
    members {
      allocation_count = 4
    }
    memory {
      allocation_mb = 2048
    }
  }
}
