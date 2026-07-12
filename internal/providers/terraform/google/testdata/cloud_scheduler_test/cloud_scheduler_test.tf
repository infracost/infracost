provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_cloud_scheduler" "non_usage" {
  name             = "example-non-usage"
  schedule         = "* * * * *"
  time_zone        = "Etc/UTC"
  http_target {
    uri = "https://example.com/task"
  }
}

resource "google_cloud_scheduler" "usage" {
  name             = "example-usage"
  schedule         = "* * * * *"
  time_zone        = "Etc/UTC"
  http_target {
    uri = "https://example.com/task"
  }
}
