provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_compute_ha_vpn_gateway" "my_compute_ha_vpn_gateway" {
  name    = "vpn1"
  network = google_compute_network.my_compute_network.id
}

resource "google_compute_network" "my_compute_network" {
  name = "network1"
}
