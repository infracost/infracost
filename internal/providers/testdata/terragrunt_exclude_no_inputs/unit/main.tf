provider "google" {
  project = "example-project"
  region  = "europe-central2"
}

resource "google_compute_address" "example" {
  name = "example"
}
