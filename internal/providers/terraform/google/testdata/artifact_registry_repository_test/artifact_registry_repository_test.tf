provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

provider "google-beta" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_artifact_registry_repository" "us_east1" {
  provider = google-beta

  location      = "us-east1"
  repository_id = "my-repository"
  format        = "DOCKER"
}

resource "google_artifact_registry_repository" "us_east1_usage" {
  provider = google-beta

  location      = "us-east1"
  repository_id = "my-repository"
  format        = "DOCKER"
}

resource "google_artifact_registry_repository" "europe_north1_usage" {
  provider = google-beta

  location      = "europe-north1"
  repository_id = "my-repository"
  format        = "DOCKER"
}

resource "google_artifact_registry_repository" "asia_east1_usage" {
  provider = google-beta

  location      = "asia-east1"
  repository_id = "my-repository"
  format        = "DOCKER"
}

resource "google_artifact_registry_repository" "australia_southeast1_usage" {
  provider = google-beta

  location      = "australia-southeast1"
  repository_id = "my-repository"
  format        = "DOCKER"
}

resource "google_artifact_registry_repository" "multiregion_europe_usage" {
  provider = google-beta

  location      = "europe"
  repository_id = "my-repository"
  format        = "DOCKER"
}
