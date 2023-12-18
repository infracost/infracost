provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_redis_instance" "basic_m1" {
  name           = "memory-cache"
  memory_size_gb = 1
}

resource "google_redis_instance" "basic_m2" {
  name           = "memory-cache"
  memory_size_gb = 5
}

resource "google_redis_instance" "basic_m3" {
  name           = "memory-cache"
  memory_size_gb = 25
}

resource "google_redis_instance" "basic_m4" {
  name           = "memory-cache"
  memory_size_gb = 45
}

resource "google_redis_instance" "basic_m5" {
  name           = "memory-cache"
  memory_size_gb = 105
}

resource "google_redis_instance" "standard_m1" {
  name           = "memory-cache"
  memory_size_gb = 1
  tier           = "STANDARD_HA"
}

resource "google_redis_instance" "standard_m2" {
  name           = "memory-cache"
  memory_size_gb = 5
  tier           = "STANDARD_HA"
}

resource "google_redis_instance" "standard_m3" {
  name           = "memory-cache"
  memory_size_gb = 25
  tier           = "STANDARD_HA"
}

resource "google_redis_instance" "standard_m4" {
  name           = "memory-cache"
  memory_size_gb = 45
  tier           = "STANDARD_HA"
}

resource "google_redis_instance" "standard_m5" {
  name           = "memory-cache"
  memory_size_gb = 105
  tier           = "STANDARD_HA"
}
