provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_compute_address" "static" {
  name    = "ipv4-address"
  purpose = "GCE_ENDPOINT"
}

resource "google_compute_address" "static_without_purpose" {
  name = "ipv4-address"
}

resource "google_compute_address" "static_with_different_purpose" {
  name    = "ipv4-address"
  purpose = "VPC_PEERING"
}

resource "google_compute_address" "internal" {
  name         = "ipv4-address-internal"
  address_type = "INTERNAL"
}

resource "google_compute_address" "default_diff_region" {
  name   = "ipv4-address-default"
  region = "europe-central2"
}

resource "google_compute_global_address" "default" {
  name = "global-appserver-ip"
}

resource "google_compute_address" "standard_static" {
  name = "ipv4-address"
}

resource "google_compute_address" "standard_with_different_purpose" {
  name    = "ipv4-address"
  purpose = "VPC_PEERING"
}

resource "google_compute_address" "preemptible_static" {
  name = "ipv4-address"
}

resource "google_compute_instance" "standard_instance_with_ip" {
  name         = "vm-instance"
  machine_type = "f1-micro"
  zone         = "us-central1-a"

  boot_disk {
    initialize_params {
      image = "centos-cloud/centos-7"
    }
  }

  network_interface {
    network = "default"
    access_config {
      nat_ip = google_compute_address.standard_static.address
    }
  }
}

resource "google_compute_instance" "standard_instance_with_ip_purpose" {
  name         = "vm-instance"
  machine_type = "f1-micro"
  zone         = "us-central1-a"

  boot_disk {
    initialize_params {
      image = "centos-cloud/centos-7"
    }
  }

  network_interface {
    network = "default"
    access_config {
      nat_ip = google_compute_address.standard_with_different_purpose.address
    }
  }
}

resource "google_compute_instance" "preemptible_instance_with_ip" {
  name         = "vm-instance"
  machine_type = "f1-micro"
  zone         = "us-central1-a"

  boot_disk {
    initialize_params {
      image = "centos-cloud/centos-7"
    }
  }

  scheduling {
    preemptible = true
  }

  network_interface {
    network = "default"
    access_config {
      nat_ip = google_compute_address.preemptible_static.address
    }
  }
}
