provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_compute_disk" "standard_default" {
  name = "standard_default"
  type = "pd-standard"
}

resource "google_compute_disk" "ssd_default" {
  name = "ssd_default"
  type = "pd-ssd"
}

resource "google_compute_disk" "extreme_default" {
  name = "extreme_default"
  type = "pd-extreme"
}

resource "google_compute_disk" "hyperdisk_default" {
  name = "hyperdisk_default"
  type = "hyperdisk-extreme"
}

resource "google_compute_disk" "size" {
  name = "size"
  type = "pd-standard"
  size = 20
}

resource "google_compute_disk" "extreme_size_iops" {
  name             = "extreme_size_iops"
  type             = "pd-extreme"
  size             = 40
  provisioned_iops = 5000
}

resource "google_compute_disk" "hyperdisk_size_iops" {
  name             = "hyperdisk_size_iops"
  type             = "hyperdisk-extreme"
  size             = 128
  provisioned_iops = 20000
}

resource "google_compute_image" "image_disk_size" {
  name         = "image_disk_size"
  disk_size_gb = 30
}

resource "google_compute_disk" "image_disk_size" {
  name  = "image_disk_size"
  type  = "pd-standard"
  image = google_compute_image.image_disk_size.self_link
}

resource "google_compute_image" "image_source_image" {
  name         = "image_source_image"
  source_image = google_compute_image.image_disk_size.self_link
}

resource "google_compute_disk" "image_source_image" {
  name  = "image_source_image"
  type  = "pd-standard"
  image = google_compute_image.image_source_image.self_link
}

resource "google_compute_snapshot" "snapshot_source_disk" {
  name        = "snapshot_source_disk"
  source_disk = google_compute_disk.size.name
}

resource "google_compute_image" "image_source_snapshot" {
  name            = "image_source_snapshot"
  source_snapshot = google_compute_snapshot.snapshot_source_disk.self_link
}

resource "google_compute_disk" "image_source_snapshot" {
  name  = "image_source_snapshot"
  type  = "pd-standard"
  image = google_compute_image.image_source_snapshot.self_link
}

resource "google_compute_disk" "snapshot_source_disk" {
  name     = "snapshot_source_disk"
  type     = "pd-standard"
  snapshot = google_compute_snapshot.snapshot_source_disk.self_link
}
