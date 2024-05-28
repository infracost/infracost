locals {
  tier1_region = "europe-west4"
  tier2_region = "europe-west3"

}
provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = local.tier2_region
}

resource "google_cloud_run_service" "throttling_enabled" {
  name     = "cloudrun-service-test"
  location = local.tier2_region
  template {
    spec {
      containers {
        image = "us-docker.pkg.dev/cloudrun/container/hello"
      }
    }
  }
}

resource "google_cloud_run_service" "throttling_disabled" {
  name     = "cloudrun-service-test"
  location = local.tier2_region
  template {
    spec {
      containers {
        image = "us-docker.pkg.dev/cloudrun/container/hello"
      }
    }
  }
  metadata {
    annotations = {
      "run.googleapis.com/cpu-throttling" : false
    }
  }
}

resource "google_cloud_run_service" "no_usage" {
  name     = "cloudrun-service-test"
  location = local.tier2_region
  template {
    spec {
      containers {
        image = "us-docker.pkg.dev/cloudrun/container/hello"
      }
    }
  }
}
