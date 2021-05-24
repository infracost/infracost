provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  region      = "us-central1"
}

resource "google_compute_address" "static" {
  name = "ipv4-address"
}

resource "google_compute_address" "internal" {
  name         = "ipv4-address-internal"
  address_type = "INTERNAL"
}

resource "google_compute_global_address" "default" {
  name = "global-appserver-ip"
}
