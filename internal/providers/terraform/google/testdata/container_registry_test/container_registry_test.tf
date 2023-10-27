provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

provider "google" {
  alias       = "asia"
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "asia-northeast1"
}

resource "google_container_registry" "my_registry" {
  project = "my-project"
}

resource "google_container_registry" "my_registry_asia" {
  provider = google.asia
  project  = "my-project-asia"
}

resource "google_container_registry" "my_registry_usage" {
  project  = "my-project"
  location = "EU"
}

