provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

# Add example resources for CloudRunService below

# resource "google_cloud_run_service" "cloud_run_service" {
# }
