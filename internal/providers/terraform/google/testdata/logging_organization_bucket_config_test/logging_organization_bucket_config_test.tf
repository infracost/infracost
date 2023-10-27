provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_logging_organization_bucket_config" "basic" {
  organization   = "1" # fake
  location       = "global"
  retention_days = 30
  bucket_id      = "_Default"
}

resource "google_logging_organization_bucket_config" "basic_withUsage" {
  organization   = "1" # fake
  location       = "global"
  retention_days = 30
  bucket_id      = "_Default"
}
