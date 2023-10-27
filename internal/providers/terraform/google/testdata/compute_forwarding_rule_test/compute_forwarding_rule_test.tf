provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_compute_forwarding_rule" "default" {
  name = "website-forwarding-rule"
}

resource "google_compute_forwarding_rule" "withoutUsage" {
  name = "website-forwarding-rule"
}

resource "google_compute_global_forwarding_rule" "default" {
  name   = "global-rule"
  target = "all-apis"
}
