provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

locals {
  node_types = ["REDIS_SHARED_CORE_NANO", "REDIS_STANDARD_SMALL", "REDIS_HIGHMEM_MEDIUM", "REDIS_HIGHMEM_XLARGE"]

  permutations = distinct(flatten([
    for node_type in local.node_types : {
      node_type = node_type
    }
  ]))
}

resource "google_compute_network" "redis_network" {
  name                    = "redis-network"
  auto_create_subnetworks = true
}

resource "google_redis_cluster" "no_aof_and_backups" {
  for_each      = { for perm in local.permutations : "${perm.node_type}" => perm }
  name          = "standard-cluster"
  region        = "us-central1"
  shard_count   = 3
  replica_count = 2

  node_type = each.value.node_type

  psc_configs {
    network = google_compute_network.redis_network.id
  }

  transit_encryption_mode = "TRANSIT_ENCRYPTION_MODE_SERVER_AUTHENTICATION"
}


resource "google_redis_cluster" "aof_and_backups" {
  name          = "basic-cluster"
  region        = "us-central1"
  shard_count   = 1
  replica_count = 2

  node_type = "REDIS_STANDARD_SMALL"

  persistence_config {
    mode = "AOF"
  }

  psc_configs {
    network = google_compute_network.redis_network.id
  }

  automated_backup_config {
    retention = 7

    fixed_frequency_schedule {
      start_time {
        hours = 1
      }
    }
  }


  transit_encryption_mode = "TRANSIT_ENCRYPTION_MODE_DISABLED"
}

resource "google_redis_cluster" "backups_with_usage" {
  name          = "basic-cluster"
  region        = "us-central1"
  shard_count   = 1
  replica_count = 0

  node_type = "REDIS_STANDARD_SMALL"
  psc_configs {
    network = google_compute_network.redis_network.id
  }

  automated_backup_config {
    retention = 7

    fixed_frequency_schedule {
      start_time {
        hours = 1
      }
    }
  }


  transit_encryption_mode = "TRANSIT_ENCRYPTION_MODE_DISABLED"
}
