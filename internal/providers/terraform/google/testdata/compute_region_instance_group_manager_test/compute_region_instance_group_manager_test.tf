provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  region      = "us-central1"
}

resource "google_compute_instance_template" "standard" {
  name = "standard-template"

  machine_type = "n1-standard-16"

  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
  }

  disk {
    source_image = "debian-cloud/debian-9"
    boot         = false
    disk_type    = "pd-balanced"
    disk_size_gb = "400"
  }

  disk {
    interface    = "NVME"
    type         = "SCRATCH"
    disk_type    = "local-ssd"
    disk_size_gb = "375"
  }
}

resource "google_compute_region_instance_group_manager" "appserver" {
  name = "appserver-igm"

  base_instance_name = "app"
  region             = "us-central1"

  version {
    instance_template = google_compute_instance_template.standard.id
  }

  target_size = 3
}