provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  region      = "us-central1"
}

resource "google_dns_managed_zone" "zone" {
  name     = "example"
  dns_name = "example-123.com."
}
