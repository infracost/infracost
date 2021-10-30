provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  region      = "us-central1"
}

resource "google_container_registry" "my_registry" {
  project = "my-project"
}

resource "google_container_registry" "my_registry_usage" {
  project  = "my-project"
  location = "EU"
}

