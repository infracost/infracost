provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_logging_project_sink" "basic" {
  name = "my-pubsub-instance-sink"

  destination = "fake"
}

resource "google_logging_project_sink" "basic_withUsage" {
  name = "my-pubsub-instance-sink"

  destination = "fake"
}
