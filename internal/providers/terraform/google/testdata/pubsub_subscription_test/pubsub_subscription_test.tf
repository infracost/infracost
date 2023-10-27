provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_pubsub_subscription" "non_usage" {
  name  = "example-subscription"
  topic = "my_topic"
}

resource "google_pubsub_subscription" "usage" {
  name  = "example-subscription"
  topic = "my_topic"
}
