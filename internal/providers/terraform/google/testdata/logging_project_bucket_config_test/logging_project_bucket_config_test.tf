provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_logging_project_bucket_config" "basic" {
  project        = "fake"
  location       = "global"
  retention_days = 30
  bucket_id      = "_Default"
}

resource "google_logging_project_bucket_config" "basic_withUsage" {
  project        = "fake"
  location       = "global"
  retention_days = 30
  bucket_id      = "_Default"
}
