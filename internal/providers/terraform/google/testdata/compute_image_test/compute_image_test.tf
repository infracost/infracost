provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_compute_image" "empty" {
  name = "example-image"
}

resource "google_compute_disk" "disk" {
  name = "test-disk"
  size = 1000
}

resource "google_compute_image" "image" {
  name         = "image-source-image"
  disk_size_gb = 100
}

resource "google_compute_snapshot" "snapshot" {
  name        = "snapshot-source-disk"
  source_disk = google_compute_disk.disk.self_link
}

resource "google_compute_image" "with_disk_size" {
  name         = "example-image"
  disk_size_gb = 500
}

resource "google_compute_image" "with_source_disk" {
  name        = "example-image"
  source_disk = google_compute_disk.disk.self_link
}

resource "google_compute_image" "with_source_image" {
  name         = "example-image"
  source_image = google_compute_image.image.self_link
}

resource "google_compute_image" "with_source_snapshot" {
  name            = "example-image"
  source_snapshot = google_compute_snapshot.snapshot.self_link
}

resource "google_compute_image" "usage" {
  name = "example-image"
}
