provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_cloud_run_service" "throttling_enabled" {
  name     = "cloud-run-service-test"
  location = "us-central1"
  template {
    spec {
      containers {
        image = "us-docker.pkg.dev/cloudrun/container/hello"
        resources {
          limits = {
            "cpu"    = "1"
            "memory" = "512Mi"
          }
        }
      }
    }
  }
  traffic {
    percent         = 100
    latest_revision = true
  }
}
resource "google_cloud_run_service" "throttling_disabled" {
  name     = "cloud-run-service-test"
  location = "us-central1"
  template {
    spec {
      containers {
        image = "us-docker.pkg.dev/cloudrun/container/hello"
      }
    }
  }
  metadata {
    annotations = {
      "run.googleapis.com/cpu-throttling": false
      "autoscaling.knative.dev/minScale": "1"
    }
  }
  traffic {
    percent         = 100
    latest_revision = true
  }
}
