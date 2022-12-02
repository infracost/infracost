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

resource "google_compute_instance" "with_hours" {
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

resource "google_compute_instance" "gpu_with_hours" {
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

resource "google_compute_instance" "custom" {
  name         = "custom"
  machine_type = "custom-6-20480"
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

resource "google_compute_instance" "custom_preemptible" {
  name         = "custom_preemptible"
  machine_type = "custom-6-20480"
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

resource "google_compute_instance" "custom_n1" {
  name         = "custom_n1"
  machine_type = "n1-custom-6-20480"
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

resource "google_compute_instance" "custom_n2" {
  name         = "custom_n2"
  machine_type = "n2-custom-6-20480"
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


resource "google_compute_instance" "custom_n2d" {
  name         = "custom_n2d"
  machine_type = "n2d-custom-6-20480"
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

resource "google_compute_instance" "custom_ext" {
  name         = "custom_ext"
  machine_type = "custom-2-15360-ext"
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

// Not supported yet
resource "google_compute_instance" "e2_custom" {
  name         = "e2_custom"
  machine_type = "e2-custom-2-15360"
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
