provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

locals {
  node_types = [
    "SHARED_CORE_NANO",
    "CUSTOM_PICO",
    "CUSTOM_MICRO",
    "CUSTOM_MINI",
    "STANDARD_SMALL",
    "HIGHMEM_MEDIUM",
    "HIGHCPU_MEDIUM",
    "STANDARD_LARGE",
    "HIGHMEM_XLARGE",
    "HIGHMEM_2XLARGE",
  ]
}

resource "google_memorystore_instance" "nodes" {
  for_each = toset(local.node_types)

  instance_id   = "valkey-${replace(lower(each.value), "_", "-")}"
  location      = "us-central1"
  shard_count   = 2
  replica_count = 1
  node_type     = each.value

  deletion_protection_enabled = false
}

resource "google_memorystore_instance" "aof_and_backups" {
  instance_id   = "valkey-aof-and-backups"
  location      = "us-central1"
  shard_count   = 2
  replica_count = 1
  node_type     = "STANDARD_SMALL"

  persistence_config {
    mode = "AOF"
  }

  automated_backup_config {
    retention = "604800s"

    fixed_frequency_schedule {
      start_time {
        hours = 1
      }
    }
  }

  deletion_protection_enabled = false
}

resource "google_memorystore_instance" "default_node_type" {
  instance_id = "valkey-default-node-type"
  location    = "us-central1"
  shard_count = 3

  deletion_protection_enabled = false
}

resource "google_memorystore_instance" "backups_with_usage" {
  instance_id = "valkey-backups-with-usage"
  location    = "us-central1"
  shard_count = 1
  node_type   = "SHARED_CORE_NANO"

  automated_backup_config {
    retention = "604800s"

    fixed_frequency_schedule {
      start_time {
        hours = 2
      }
    }
  }

  deletion_protection_enabled = false
}
