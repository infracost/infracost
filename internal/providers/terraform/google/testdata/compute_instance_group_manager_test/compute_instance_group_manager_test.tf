provider "google" {
  credentials = "{\"type\":\"service_account\"}"
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

  disk {
    interface    = "NVME"
    type         = "SCRATCH"
    disk_type    = "local-ssd"
    disk_size_gb = "375"
  }

  guest_accelerator {
    type  = "nvidia-tesla-k80"
    count = 2
  }
}

resource "google_compute_instance_group_manager" "default" {
  name = "default"

  base_instance_name = "app"
  zone               = "us-central1-a"

  version {
    instance_template = google_compute_instance_template.appserver.id
  }

  target_size = 4
}
