provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_compute_network" "redis_network" {
  name                    = "redis-network"
  auto_create_subnetworks = true
}

# Basic cluster - AOF enabled, backups enabled (removed unsupported memory_size_gb and automated_backup_config)
resource "google_redis_cluster" "basic_cluster" {
  name          = "basic-cluster"
  region        = "us-central1"
  shard_count   = 1
  replica_count = 2

  node_type      = "REDIS_STANDARD_SMALL"
  
  persistence_config {
    mode = "AOF"
  }

  psc_configs {
    network = google_compute_network.redis_network.id
  }

  transit_encryption_mode = "TRANSIT_ENCRYPTION_MODE_DISABLED"
}

# Standard cluster - No AOF, no backups (testing default behavior)
resource "google_redis_cluster" "standard_cluster" {
  name          = "standard-cluster"
  region        = "us-central1"
  shard_count   = 3
  replica_count = 2

  node_type      = "REDIS_HIGHMEM_XLARGE"
  
  psc_configs {
    network = google_compute_network.redis_network.id
  }

  transit_encryption_mode = "TRANSIT_ENCRYPTION_MODE_SERVER_AUTHENTICATION"
}

# Basic cluster with zero replicas - minimal setup, no AOF, no backups
resource "google_redis_cluster" "basic_cluster_zero" {
  name          = "basic-cluster-zero"
  region        = "us-central1"
  shard_count   = 1
  replica_count = 0

  node_type      = "REDIS_STANDARD_SMALL"
  
  psc_configs {
    network = google_compute_network.redis_network.id
  }

  transit_encryption_mode = "TRANSIT_ENCRYPTION_MODE_DISABLED"
}

resource "google_redis_cluster" "cluster_with_backups" {
  name          = "cluster-with-backups"
  region        = "us-central1"
  shard_count   = 1
  replica_count = 2
  node_type      = "REDIS_STANDARD_SMALL"

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
