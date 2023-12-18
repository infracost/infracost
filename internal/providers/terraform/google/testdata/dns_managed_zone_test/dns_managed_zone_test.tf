provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_dns_managed_zone" "zone" {
  name     = "example"
  dns_name = "example-123.com."
}
