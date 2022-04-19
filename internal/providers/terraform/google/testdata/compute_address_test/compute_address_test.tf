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

resource "google_compute_address" "standard_with_usage" {
  name = "ipv4-address"
}

resource "google_compute_address" "preemptible_with_usage" {
  name = "ipv4-address"
}

resource "google_compute_address" "unused_with_usage" {
  name = "ipv4-address"
}

resource "google_compute_address" "invalid_usage" {
  name = "ipv4-address"
}

resource "google_compute_global_address" "standard_with_usage" {
  name = "global-appserver-ip"
}
