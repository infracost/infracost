provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_compute_network" "redis_network" {
  name                    = "redis-network"
  auto_create_subnetworks = true
}

resource "google_redis_cluster" "basic_cluster" {
  name          = "basic-cluster"
  region        = "us-central1"
  shard_count   = 1
  replica_count = 1

  psc_configs {
    network = google_compute_network.redis_network.id
  }

  transit_encryption_mode = "TRANSIT_ENCRYPTION_MODE_DISABLED"
}

resource "google_redis_cluster" "standard_cluster" {
  name          = "standard-cluster"
  region        = "us-central1"
  shard_count   = 3
  replica_count = 2

  psc_configs {
    network = google_compute_network.redis_network.id
  }

  transit_encryption_mode = "TRANSIT_ENCRYPTION_MODE_SERVER_AUTHENTICATION"
}
