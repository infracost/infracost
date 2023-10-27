provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_logging_folder_sink" "basic" {
  name        = "my-sink"
  description = "what it is"
  folder      = "fake"

  destination = "storage.googleapis.com/fake"
}

resource "google_logging_folder_sink" "basic_withUsage" {
  name        = "my-sink"
  description = "what it is"
  folder      = "fake"

  destination = "storage.googleapis.com/fake"
}
