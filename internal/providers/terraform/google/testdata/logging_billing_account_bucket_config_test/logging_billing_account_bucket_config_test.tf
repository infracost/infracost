provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_logging_billing_account_bucket_config" "basic" {
  billing_account = "00AA00-000AAA-00AA0A" # fake
  location        = "global"
  retention_days  = 30
  bucket_id       = "_Default"
}

resource "google_logging_billing_account_bucket_config" "basic_withUsage" {
  billing_account = "00AA00-000AAA-00AA0A" # fake
  location        = "global"
  retention_days  = 30
  bucket_id       = "_Default"
}
