provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
}

resource "google_cloud_run_v2_service" "throttling_enabled" {
  name     = "cloudrun-service-test-with-cpu-idle"
  location = "europe-west3"
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    containers {
      image = "us-docker.pkg.dev/cloudrun/container/hello"
    }
  }
}

resource "google_cloud_run_v2_service" "throttling_disabled" {
  name     = "cloudrun-service-test-with-cpu-idle"
  location = "europe-west3"
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    containers {
      image = "us-docker.pkg.dev/cloudrun/container/hello"
      resources {
        limits = {
          cpu    = "1"
          memory = "512Mi"
        }
        cpu_idle = false
      }
    }
    scaling {
      min_instance_count = 2
    }
  }
}

resource "google_cloud_run_v2_service" "throttling_enabled_no_usage" {
  name     = "cloudrun-service-test-with-cpu-idle"
  location = "europe-west3"
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    containers {
      image = "us-docker.pkg.dev/cloudrun/container/hello"
    }
  }
}
