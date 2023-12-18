provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_secret_manager_secret" "secret_example" {
  secret_id = "secret"

  labels = {
    label = "example"
  }

  replication {
    user_managed {
      replicas {
        location = "us-central1"
      }
      replicas {
        location = "us-east1"
      }
    }
  }
}

resource "google_secret_manager_secret" "secret_automatic" {
  secret_id = "secret-with-usage"

  labels = {
    label = "example-with-usage"
  }

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "secret_version" {
  secret = google_secret_manager_secret.secret_example.id

  secret_data = "secret-data"
}

resource "google_secret_manager_secret_version" "secret_version_with_usage" {
  secret = google_secret_manager_secret.secret_automatic.id

  secret_data = "secret-data"
}

resource "google_secret_manager_secret_version" "secret_version_replicas_with_usage" {
  secret = google_secret_manager_secret.secret_example.id

  secret_data = "secret-data"
}
