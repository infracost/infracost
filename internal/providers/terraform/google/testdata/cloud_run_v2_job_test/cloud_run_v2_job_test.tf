provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "europe-west3"
}

resource "google_cloud_run_v2_job" "basic" {
  name = "cloudrun-v2-job-test"
  location = "europe-west4"
  template {
    template {
      containers {
        image = "us-docker.pkg.dev/cloudrun/container/hello"
      }
    }
    task_count = 1
  }
}
