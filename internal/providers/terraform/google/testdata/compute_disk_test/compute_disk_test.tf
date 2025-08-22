provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_compute_disk" "standard_default" {
  name = "standard-default"
  type = "pd-standard"
}

resource "google_compute_disk" "ssd_default" {
  name = "ssd-default"
  type = "pd-ssd"
}

resource "google_compute_disk" "extreme_default" {
  name = "extreme-default"
  type = "pd-extreme"
}

resource "google_compute_disk" "hyperdisk_default" {
  name = "hyperdisk-default"
  type = "hyperdisk-extreme"
}

resource "google_compute_disk" "size" {
  name = "size"
  type = "pd-standard"
  size = 20
}

resource "google_compute_disk" "extreme_size_iops" {
  name             = "extreme-size-iops"
  type             = "pd-extreme"
  size             = 40
  provisioned_iops = 5000
}

resource "google_compute_disk" "hyperdisk_size_iops" {
  name             = "hyperdisk-size-iops"
  type             = "hyperdisk-extreme"
  size             = 128
  provisioned_iops = 20000
}

resource "google_compute_image" "image_disk_size" {
  name         = "image-disk-size"
  disk_size_gb = 30
}

resource "google_compute_disk" "image_disk_size" {
  name  = "image-disk-size"
  type  = "pd-standard"
  image = google_compute_image.image_disk_size.self_link
}

resource "google_compute_image" "image_source_image" {
  name         = "image-source-image"
  source_image = google_compute_image.image_disk_size.self_link
}

resource "google_compute_disk" "image_source_image" {
  name  = "image-source-image"
  type  = "pd-standard"
  image = google_compute_image.image_source_image.self_link
}

resource "google_compute_snapshot" "snapshot_source_disk" {
  name        = "snapshot-source-disk"
  source_disk = google_compute_disk.size.name
}

resource "google_compute_image" "image_source_snapshot" {
  name            = "image-source-snapshot"
  source_snapshot = google_compute_snapshot.snapshot_source_disk.self_link
}

resource "google_compute_disk" "image_source_snapshot" {
  name  = "image-source-snapshot"
  type  = "pd-standard"
  image = google_compute_image.image_source_snapshot.self_link
}

resource "google_compute_disk" "snapshot_source_disk" {
  name     = "snapshot-source-disk"
  type     = "pd-standard"
  snapshot = google_compute_snapshot.snapshot_source_disk.self_link
}
