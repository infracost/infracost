provider "google-beta" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_compute_instance" "vm" {
  name         = "vm"
  machine_type = "e2-medium"
  zone         = "us-central1-a"

  boot_disk {
    initialize_params {
      image = "fake"
    }
  }

  network_interface {
    network = "fake"
  }
}

resource "google_compute_machine_image" "image" {
  provider        = "google-beta"
  name            = "image"
  source_instance = google_compute_instance.vm.self_link
}

resource "google_compute_machine_image" "usage" {
  provider        = "google-beta"
  name            = "image"
  source_instance = google_compute_instance.vm.self_link
}
