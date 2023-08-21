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
  machine_type = "n2d-custom-4-20480"
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

// Sustained use discount TWENTY Percentage
resource "google_compute_instance" "n2_standard_8" {
  name         = "n2_standard_8"
  machine_type = "n2-standard-8"
  zone         = "us-central1"

  usage_period {
    start_time = "2023-08-01T00:00:00Z" // Specify the start time of the discount usage period
    end_time   = "2023-08-30T23:59:59Z" // Specify the end time of the discount usage period
    usage      = "470"                  // Specify the monthly hours value (e.g., 730 for 730 hours)
  }

  boot_disk {
    initialize_params {
      image = "centos-cloud/centos-7"
    }
  }

  network_interface {
    network = "default"
  }
}

// Sustained use discount THIRTY Percentage
resource "google_compute_instance" "m1_ultramem_80" {
  name         = "m1_ultramem_80"
  machine_type = "m1-ultramem-80"
  zone         = "us-central1"

  usage_period {
    start_time = "2023-08-01T00:00:01Z" // Specify the start time of the discount usage period
    end_time   = "2023-08-30T23:59:59Z" // Specify the end time of the discount usage period
    usage      = "580"                  // Specify the monthly hours value (e.g., 730 for 730 hours)
  }

  boot_disk {
    initialize_params {
      image = "centos-cloud/centos-7"
    }
  }

  network_interface {
    network = "default"
  }
}

// Sustained use discount UNSUPPORTED TYPE will set the SUD to ZERO
resource "google_compute_instance" "t2d_standard_4" {
  name         = "t2d_standard_4"
  machine_type = "t2d-standard-4"
  zone         = "us-central1"

  usage_period {
    start_time = "2023-08-01T00:00:01Z" // Specify the start time of the discount usage period
    end_time   = "2023-08-30T23:59:59Z" // Specify the end time of the discount usage period
    usage      = "390"                  // Specify the monthly hours value (e.g., 730 for 730 hours)
  }

  boot_disk {
    initialize_params {
      image = "centos-cloud/centos-7"
    }
  }

  network_interface {
    network = "default"
  }
}
