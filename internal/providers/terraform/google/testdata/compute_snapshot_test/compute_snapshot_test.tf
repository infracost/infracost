provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  region      = "us-central1"
}

resource "google_compute_disk" "default" {
  name = "test-disk"
  size = 100
}

resource "google_compute_snapshot" "snapshot" {
  name        = "my-snapshot"
  source_disk = google_compute_disk.default.name
}
