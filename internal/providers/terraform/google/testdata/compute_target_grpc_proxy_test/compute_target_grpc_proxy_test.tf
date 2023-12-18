provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_compute_target_grpc_proxy" "default" {
  name = "proxy"
}
resource "google_compute_target_grpc_proxy" "withoutUsage" {
  name = "proxy"
}

resource "google_compute_target_http_proxy" "default" {
  name    = "proxy"
  url_map = "fake"
}

resource "google_compute_target_https_proxy" "default" {
  name             = "proxy"
  url_map          = "fake"
  ssl_certificates = ["123.123.123.123]"]
}

resource "google_compute_target_ssl_proxy" "default" {
  name             = "proxy"
  backend_service  = "fake"
  ssl_certificates = ["123.123.123.123]"]
}

resource "google_compute_target_tcp_proxy" "default" {
  name            = "proxy"
  backend_service = "fake"
}

resource "google_compute_region_target_http_proxy" "default" {
  name    = "proxy"
  url_map = "fake"
}

resource "google_compute_region_target_https_proxy" "default" {
  name             = "proxy"
  ssl_certificates = ["123.123.123.123]"]
  url_map          = "fake"
}
