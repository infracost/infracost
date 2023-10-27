provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_bigquery_dataset" "usage" {
  dataset_id  = "example_dataset"
  description = "This is a test description"
}

resource "google_bigquery_dataset" "non_usage" {
  dataset_id  = "example_dataset"
  description = "This is a test description"
}
