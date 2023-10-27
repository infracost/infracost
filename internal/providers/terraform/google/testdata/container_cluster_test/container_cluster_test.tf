provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_container_cluster" "zonal" {
  name     = "zonal"
  location = "us-central1-a"
}

resource "google_container_cluster" "regional" {
  name     = "regional"
  location = "us-central1"
}

resource "google_container_cluster" "node_locations" {
  name     = "node-locations"
  location = "us-central1"

  node_locations = [
    "us-central1-a",
    "us-central1-b"
  ]
}

resource "google_container_cluster" "with_node_config" {
  name               = "with-node-config"
  location           = "us-central1-a"
  initial_node_count = 3

  node_config {
    machine_type    = "n1-standard-16"
    disk_size_gb    = 120
    disk_type       = "pd-ssd"
    local_ssd_count = 1

    guest_accelerator {
      type  = "nvidia-tesla-k80"
      count = 4
    }
  }
}

resource "google_container_cluster" "with_node_pools_zonal" {
  name     = "with-node-pools"
  location = "us-central1-a"

  node_pool {
    node_count = 2

    node_config {
      machine_type = "n1-standard-16"
    }
  }

  node_pool {
    node_count = 4

    node_config {
      machine_type = "n1-standard-16"
      preemptible  = true
    }
  }
}

resource "google_container_cluster" "with_node_pools_regional" {
  name     = "with-node-pools-regional"
  location = "us-central1"

  node_pool {
    node_count = 2

    node_config {
      machine_type = "n1-standard-16"
    }
  }

  node_pool {
    node_count = 4

    node_config {
      machine_type = "n1-standard-16"
      preemptible  = true
    }
  }
}

resource "google_container_cluster" "with_node_pools_node_locations" {
  name     = "with-node-pools-regional"
  location = "us-central1"

  node_locations = [
    "us-central1-a",
    "us-central1-b"
  ]

  node_pool {
    node_count = 2

    node_config {
      machine_type = "n1-standard-16"
    }
  }

  node_pool {
    node_count = 2

    node_locations = [
      "us-central1-a"
    ]

    node_config {
      machine_type = "n1-standard-16"
      preemptible  = true
    }
  }
}

resource "google_container_cluster" "zonal_withUsage" {
  name               = "zonal"
  location           = "us-central1-a"
  initial_node_count = 3
}

resource "google_container_cluster" "regional_withUsage" {
  name               = "regional"
  location           = "us-central1"
  initial_node_count = 3
}

resource "google_container_cluster" "node_locations_withUsage" {
  name               = "node-locations"
  location           = "us-central1"
  initial_node_count = 3

  node_locations = [
    "us-central1-a",
    "us-central1-b"
  ]
}

resource "google_container_cluster" "with_node_pools_zonal_withUsage" {
  name     = "with-node-pools"
  location = "us-central1-a"

  node_pool {
    node_count = 2

    node_config {
      machine_type = "n1-standard-16"
    }
  }

  node_pool {
    node_count = 4

    node_config {
      machine_type = "n1-standard-16"
      preemptible  = true
    }
  }
}

resource "google_container_cluster" "with_node_pools_regional_withUsage" {
  name     = "with-node-pools-regional"
  location = "us-central1"

  node_pool {
    node_count = 2

    node_config {
      machine_type = "n1-standard-16"
    }
  }

  node_pool {
    node_count = 4

    node_config {
      machine_type = "n1-standard-16"
      preemptible  = true
    }
  }
}

resource "google_container_cluster" "with_unsupported_node_pool" {
  name     = "with-node-pools-regional"
  location = "us-central1"

  node_pool {
    node_count = 2

    node_config {
      machine_type = "e2-custom"
    }
  }

  node_pool {
    node_count = 4

    node_config {
      machine_type = "n1-standard-16"
      preemptible  = true
    }
  }
}


resource "google_container_cluster" "with_node_pools_node_locations_withUsage" {
  name     = "with-node-pools-regional"
  location = "us-central1"

  node_locations = [
    "us-central1-a",
    "us-central1-b"
  ]

  node_pool {
    node_count = 2

    node_config {
      machine_type = "n1-standard-16"
    }
  }

  node_pool {
    node_count = 2

    node_locations = [
      "us-central1-a"
    ]

    node_config {
      machine_type = "n1-standard-16"
      preemptible  = true
    }
  }
}

resource "google_container_cluster" "autopilot" {
  name     = "autopilot"
  location = "us-central1"

  enable_autopilot = true
}

resource "google_container_cluster" "autopilot_with_usage" {
  name     = "autopilot-with-usage"
  location = "us-central1"

  enable_autopilot = true
}
