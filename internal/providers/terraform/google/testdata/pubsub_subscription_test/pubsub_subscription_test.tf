provider "google" {
  credentials = "{\"type\":\"service_account\"}"
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
