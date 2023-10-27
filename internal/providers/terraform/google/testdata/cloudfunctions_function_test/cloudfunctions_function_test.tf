provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_cloudfunctions_function" "function" {
  name        = "function-test"
  description = "My function"
  runtime     = "nodejs10"
}

resource "google_cloudfunctions_function" "my_function" {
  name                = "function-test"
  description         = "My function"
  runtime             = "nodejs10"
  available_memory_mb = 256
}
