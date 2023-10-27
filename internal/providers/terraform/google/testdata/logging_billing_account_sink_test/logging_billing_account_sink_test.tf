provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_logging_billing_account_sink" "basic" {
  name            = "my-sink"
  description     = "what it is"
  billing_account = "00AA00-000AAA-00AA0A" # fake

  destination = "storage.googleapis.com/fake"
}

resource "google_logging_billing_account_sink" "my-sink_withUsage" {
  name            = "my-sink"
  description     = "what it is"
  billing_account = "00AA00-000AAA-00AA0A" # fake

  destination = "storage.googleapis.com/fake"
}
