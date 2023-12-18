provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_compute_instance_template" "appserver" {
  name = "appserver-template"

  machine_type = "f1-micro"

  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
  }

  disk {
    source_image = "debian-cloud/debian-9"
    boot         = true
    disk_type    = "pd-ssd"
    disk_size_gb = "375"
  }

}

resource "google_compute_region_instance_group_manager" "appserver" {
  name = "appserver"

  base_instance_name = "appserver"
  region             = "us-central1"

  version {
    instance_template = google_compute_instance_template.appserver.id
  }

  target_size = 0
}

resource "google_compute_region_per_instance_config" "appserver" {
  name   = "instance-1"
  region = google_compute_region_instance_group_manager.appserver.region

  region_instance_group_manager = google_compute_region_instance_group_manager.appserver.name

  preserved_state {
    metadata = {
      instance_template = google_compute_instance_template.appserver.id
    }
  }
}
