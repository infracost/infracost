provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_dns_record_set" "frontend" {
  name         = "frontend.123."
  type         = "A"
  ttl          = 300
  rrdatas      = ["123.123.123.123]"]
  managed_zone = "zone"
}

resource "google_dns_record_set" "frontend_usage" {
  name         = "frontend.123."
  type         = "A"
  ttl          = 300
  rrdatas      = ["123.123.123.123]"]
  managed_zone = "zone"
}
