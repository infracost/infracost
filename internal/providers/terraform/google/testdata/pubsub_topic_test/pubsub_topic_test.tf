provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_pubsub_topic" "non_usage" {
  name = "example-topic"
}

resource "google_pubsub_topic" "usage" {
  name = "example-topic"
}
