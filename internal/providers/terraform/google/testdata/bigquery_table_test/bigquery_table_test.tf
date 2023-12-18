provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_bigquery_dataset" "default" {
  dataset_id = "foo"
}

resource "google_bigquery_table" "usage" {
  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = "bar"
}

resource "google_bigquery_table" "non_usage" {
  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = "bar"
}
