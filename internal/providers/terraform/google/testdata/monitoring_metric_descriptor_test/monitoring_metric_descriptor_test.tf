provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_monitoring_metric_descriptor" "non_usage" {
  description  = "Daily sales records from all branch stores."
  display_name = "metric-descriptor"
  type         = "custom.googleapis.com/stores/daily_sales"
  metric_kind  = "GAUGE"
  value_type   = "DOUBLE"
}

resource "google_monitoring_metric_descriptor" "usage" {
  description  = "Daily sales records from all branch stores."
  display_name = "metric-descriptor"
  type         = "custom.googleapis.com/stores/daily_sales"
  metric_kind  = "GAUGE"
  value_type   = "DOUBLE"
}
