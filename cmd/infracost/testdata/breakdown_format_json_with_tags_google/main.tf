provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  region      = "us-central1"
}

resource "google_compute_disk" "gcd1" {
  name = "gcd1"
  type = "pd-standard"

  labels = {
    GoogleLabel = "compute-disk-label"
  }
}

resource "google_compute_disk" "gcd2" {
  name = "gcd2"
  type = "pd-ssd"
}
