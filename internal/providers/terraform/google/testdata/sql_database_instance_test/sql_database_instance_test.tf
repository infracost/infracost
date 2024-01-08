provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

locals {
  legacy_tiers      = ["db-n1-standard-1", "db-n1-standard-2", "db-n1-standard-4", "db-n1-standard-8", "db-n1-standard-16", "db-n1-standard-32", "db-n1-standard-64", "db-n1-standard-96"]
  custom_tiers      = ["db-custom-1-3840", "db-custom-2-7680", "db-custom-4-15360", "db-custom-8-30720", "db-custom-16-61440", "db-custom-32-122880", "db-custom-64-245760", "db-custom-96-368640"]
  current_tiers     = ["db-f1-micro", "db-g1-small", "db-lightweight-1", "db-lightweight-2", "db-lightweight-4", "db-standard-1", "db-standard-2", "db-standard-4", "db-highmem-4", "db-highmem-8", "db-highmem-16"]
  database_versions = ["MYSQL_8_0", "POSTGRES_15", "SQLSERVER_2019_WEB"]

  availability_types = ["ZONAL", "REGIONAL"]

  all_tiers = concat(local.legacy_tiers, local.custom_tiers, local.current_tiers)

  // Skip SQL Server on db-f1-micro and db-g1-small since that's not valid
  permutations = distinct(flatten([
    for tier in local.all_tiers : [
      for database_version in local.database_versions : [
        for availability_type in local.availability_types : {
          tier              = tier
          database_version  = database_version
          availability_type = availability_type
        }
      ] if !(contains(["db-f1-micro", "db-g1-small"], tier) && database_version == "SQLSERVER_2019_WEB")
    ]
  ]))
}

resource "google_sql_database_instance" "db_instance" {
  for_each = { for entry in local.permutations : "${entry.tier}.${entry.database_version}.${entry.availability_type}" => entry }

  name             = "db-instance-${each.value.tier}-${each.value.database_version}-${each.value.availability_type}"
  database_version = each.value.database_version

  settings {
    tier              = each.value.tier
    availability_type = each.value.availability_type
  }
}

resource "google_sql_database_instance" "no_public_ip" {
  name             = "no-public-ip"
  database_version = "MYSQL_8_0"

  settings {
    tier              = "db-standard-1"
    availability_type = "ZONAL"

    ip_configuration {
      ipv4_enabled = false
    }
  }
}

resource "google_sql_database_instance" "with_replica" {
  name             = "with-replica"
  database_version = "MYSQL_8_0"
  settings {
    tier              = "db-standard-1"
    availability_type = "REGIONAL"
    disk_size         = 500
  }
  replica_configuration {
    username = "replica"
  }
}


resource "google_sql_database_instance" "enterprise_plus" {
  name             = "with-replica"
  database_version = "MYSQL_8_0"
  settings {
    tier              = "db-standard-1"
    availability_type = "REGIONAL"
    edition           = "ENTERPRISE_PLUS"
  }
}

resource "google_sql_database_instance" "usage" {
  name             = "usage"
  database_version = "MYSQL_8_0"

  settings {
    tier              = "db-g1-small"
    availability_type = "ZONAL"
  }
}
