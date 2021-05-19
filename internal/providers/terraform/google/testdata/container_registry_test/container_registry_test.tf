provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  region      = "us-central1"
}

resource "google_container_registry" "my_registry" {
  project  = "my-project"
  location = "EU"
}
