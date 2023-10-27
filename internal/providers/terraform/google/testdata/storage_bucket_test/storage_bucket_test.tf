provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_storage_bucket" "storage_bucket" {
  name          = "auto-expiring-bucket"
  location      = "ASIA"
  storage_class = "COLDLINE"
  force_destroy = true

  lifecycle_rule {
    condition {
      age = 3
    }
    action {
      type = "Delete"
    }
  }
}

resource "google_storage_bucket" "EuMulti" {
  name          = "test"
  location      = "EU"
  force_destroy = false
}

resource "google_storage_bucket" "non_usage" {
  name          = "test"
  location      = "EU"
  force_destroy = false
}
