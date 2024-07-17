provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
  default_labels = {
    DefaultLabel = "this is a default label"
  }
}

resource "google_compute_disk" "gcd1" {
  name = "gcd1"
  type = "pd-standard"

  labels = {
    GoogleLabel = "compute-disk-label"
  }
}

resource "google_compute_disk" "gcd2" {
  name = "gcd2"
  type = "pd-ssd"
}

resource "google_monitoring_custom_service" "gmcs" {
  service_id = "custom-srv"

  user_labels = {
    GoogleUserLabel = "monitoring-custom-service-label"
  }
}

resource "google_sql_database_instance" "gsdi" {
  name             = "main-instance"
  database_version = "POSTGRES_15"

  settings {
    tier = "db-f1-micro"
    user_labels = {
      GoogleSettingsUserLabel = "sql-db-label"
    }
  }
}

