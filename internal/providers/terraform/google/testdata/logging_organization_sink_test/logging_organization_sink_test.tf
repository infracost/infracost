provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_logging_organization_sink" "basic" {
  name        = "basic"
  description = "what it is"
  org_id      = "fake"

  destination = "storage.googleapis.com/fake"
}

resource "google_logging_organization_sink" "basic_withUsage" {
  name        = "basic"
  description = "what it is"
  org_id      = "fake"

  destination = "storage.googleapis.com/fake"
}
