provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
}

resource "google_cloud_run_v2_job" "my_job" {
  name     = "cloudrun-v2-job-test"
  location = "europe-west3"
  template {
    template {
      containers {
        image = "us-docker.pkg.dev/cloudrun/container/hello"
      }
    }
    task_count = 1
  }
}

resource "google_cloud_run_v2_job" "no_usage" {
  name     = "cloudrun-v2-job-test"
  location = "europe-west3"
  template {
    template {
      containers {
        image = "us-docker.pkg.dev/cloudrun/container/hello"
      }
    }
    task_count = 1
  }
}
