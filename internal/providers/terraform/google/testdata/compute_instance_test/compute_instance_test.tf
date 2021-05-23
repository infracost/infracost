provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  region      = "us-central1"
}

resource "google_compute_instance" "standard" {
  name         = "standard"
  machine_type = "f1-micro"
  zone         = "us-central1-a"

  boot_disk {
    initialize_params {
      image = "centos-cloud/centos-7"
    }
  }

  network_interface {
    network = "default"
  }
}

resource "google_compute_instance" "ssd" {
  name         = "ssd"
  machine_type = "f1-micro"
  zone         = "us-central1-a"

  boot_disk {
    initialize_params {
      image = "centos-cloud/centos-7"
      size  = 40
      type  = "pd-ssd"
    }
  }

  network_interface {
    network = "default"
  }
}

resource "google_compute_instance" "preemptible" {
  name         = "preemptible"
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
  }
}

resource "google_compute_instance" "local_ssd" {
  name         = "local_ssd"
  machine_type = "f1-micro"
  zone         = "us-central1-a"

  boot_disk {
    initialize_params {
      image = "centos-cloud/centos-7"
    }
  }

  network_interface {
    network = "default"
  }

  scratch_disk {
    interface = "SCSI"
  }

  scratch_disk {
    interface = "SCSI"
  }
}

resource "google_compute_instance" "preemptible_local_ssd" {
  name         = "preemptible_local_ssd"
  machine_type = "f1-micro"
  zone         = "us-central1-a"

  boot_disk {
    initialize_params {
      image = "centos-cloud/centos-7"
    }
  }

  network_interface {
    network = "default"
  }

  scheduling {
    preemptible = true
  }

  scratch_disk {
    interface = "SCSI"
  }
}


resource "google_compute_instance" "gpu" {
  name         = "gpu"
  machine_type = "n1-standard-16"
  zone         = "us-central1-a"

  boot_disk {
    initialize_params {
      image = "centos-cloud/centos-7"
    }
  }

  guest_accelerator {
    type  = "nvidia-tesla-k80"
    count = 4
  }

  network_interface {
    network = "default"
  }
}

resource "google_compute_instance" "preemptible_gpu" {
  name         = "preemptible_gpu"
  machine_type = "n1-standard-16"
  zone         = "us-central1-a"

  boot_disk {
    initialize_params {
      image = "centos-cloud/centos-7"
    }
  }

  guest_accelerator {
    type  = "nvidia-tesla-k80"
    count = 4
  }

  scheduling {
    preemptible = true
  }

  network_interface {
    network = "default"
  }
}
