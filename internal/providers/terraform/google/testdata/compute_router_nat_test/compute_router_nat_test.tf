provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_compute_router_nat" "no_usage" {
  name                               = "example"
  router                             = "example-router"
  region                             = "us-central1"
  nat_ip_allocate_option             = "MANUAL_ONLY"
  source_subnetwork_ip_ranges_to_nat = "LIST_OF_SUBNETWORKS"
}

resource "google_compute_router_nat" "nat" {
  name                               = "example"
  router                             = "example-router"
  region                             = "us-central1"
  nat_ip_allocate_option             = "MANUAL_ONLY"
  source_subnetwork_ip_ranges_to_nat = "LIST_OF_SUBNETWORKS"
}

resource "google_compute_router_nat" "over_32_vms" {
  name                               = "example-over-32-vms"
  router                             = "example-router"
  region                             = "us-central1"
  nat_ip_allocate_option             = "MANUAL_ONLY"
  source_subnetwork_ip_ranges_to_nat = "LIST_OF_SUBNETWORKS"
}
