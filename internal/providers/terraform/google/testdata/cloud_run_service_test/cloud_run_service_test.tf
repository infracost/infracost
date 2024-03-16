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
  name     = "cloud-run-service-test"
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
      "autoscaling.knative.dev/minScale": "1"
    }
  }
  traffic {
    percent         = 100
    latest_revision = true
  }
}
resource "google_cloud_run_service" "throttling_disabled" {
  name     = "cloud-run-service-test"
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
      "run.googleapis.com/cpu-throttling": false
    }
  }
  traffic {
    percent         = 100
    latest_revision = true
  }
}
# resource "google_cloud_run_service" "throttling_disabled_with_resource_limits" {
#   name     = "cloud-run-service-test"
#   location = local.tier2_region
#   template {
#     spec {
#       containers {
#         image = "us-docker.pkg.dev/cloudrun/container/hello"
#         resources {
#           limits = {
#             "cpu"    = "1"
#             "memory" = "512Mi"
#           }
#         }
#       }
#     }
#   }
#   metadata {
#     annotations = {
#       "run.googleapis.com/cpu-throttling": false
#     }
#   }
#   traffic {
#     percent         = 100
#     latest_revision = true
#   }
# }
# resource "google_cloud_run_service" "throttling_disabled_with_autoscaling" {
#   name     = "cloud-run-service-test"
#   location = local.tier2_region
#   template {
#     spec {
#       containers {
#         image = "us-docker.pkg.dev/cloudrun/container/hello"
#       }
#     }
#   }
#   metadata {
#     annotations = {
#       "run.googleapis.com/cpu-throttling": false
#       "autoscaling.knative.dev/minScale": "1"
#     }
#   }
#   traffic {
#     percent         = 100
#     latest_revision = true
#   }
# }
