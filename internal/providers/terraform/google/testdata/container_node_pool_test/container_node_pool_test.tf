provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_container_cluster" "default_regional" {
  name     = "default"
  location = "us-central1-a"
}

resource "google_container_cluster" "node_locations_regional" {
  name     = "node-locations"
  location = "us-central1-a"

  node_locations = [
    "us-central1-a",
    "us-central1-b",
  ]
}

resource "google_container_node_pool" "default_regional" {
  name    = "default"
  cluster = google_container_cluster.default_regional.id
}

resource "google_container_node_pool" "with_node_config_regional" {
  name       = "with-node-config"
  cluster    = google_container_cluster.default_regional.id
  node_count = 3

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

resource "google_container_node_pool" "with_custom_instance" {
  name       = "with-custom-instance"
  cluster    = google_container_cluster.default_regional.id
  node_count = 3

  node_config {
    machine_type = "n1-custom-6-20480"
  }
}

resource "google_container_node_pool" "with_preemptible_instance" {
  name       = "with-preemptible-instance"
  cluster    = google_container_cluster.default_regional.id
  node_count = 3

  node_config {
    machine_type = "n1-custom-6-20480"
    preemptible  = true
  }

}

resource "google_container_node_pool" "with_spot_instance" {
  name       = "with-preemptible-instance"
  cluster    = google_container_cluster.default_regional.id
  node_count = 3

  node_config {
    machine_type = "n1-custom-6-20480"
    spot         = true
  }

}

resource "google_container_node_pool" "with_guest_accelerator_a100" {
  name       = "with-a100"
  cluster    = google_container_cluster.default_regional.id
  node_count = 3

  node_config {
    machine_type    = "n1-standard-16"
    disk_size_gb    = 120
    disk_type       = "pd-ssd"
    local_ssd_count = 1

    guest_accelerator {
      type  = "nvidia-tesla-a100"
      count = 4
    }
  }
}

resource "google_container_node_pool" "cluster_node_locations_regional" {
  name       = "cluster-node-locations"
  cluster    = google_container_cluster.node_locations_regional.id
  node_count = 2
}

resource "google_container_node_pool" "node_locations_regional" {
  name       = "node-locations"
  cluster    = google_container_cluster.default_regional.id
  node_count = 2

  node_locations = [
    "us-central1-a",
    "us-central1-b",
  ]
}

resource "google_container_node_pool" "initial_node_count_regional" {
  name               = "initial-node-count"
  cluster            = google_container_cluster.default_regional.id
  initial_node_count = 4
}

resource "google_container_node_pool" "autoscaling_regional" {
  name    = "autoscaling"
  cluster = google_container_cluster.default_regional.id

  autoscaling {
    min_node_count = 2
    max_node_count = 10
  }
}

resource "google_container_cluster" "default_zonal" {
  name     = "default"
  location = "us-central1"
}

resource "google_container_cluster" "node_locations_zonal" {
  name     = "node-locations"
  location = "us-central1"

  node_locations = [
    "us-central1-a",
    "us-central1-b",
  ]
}

resource "google_container_node_pool" "default_zonal" {
  name    = "default"
  cluster = google_container_cluster.default_zonal.id
}

resource "google_container_node_pool" "cluster_node_locations_zonal" {
  name       = "cluster-node-locations"
  cluster    = google_container_cluster.node_locations_zonal.id
  node_count = 2
}

resource "google_container_node_pool" "node_locations_zonal" {
  name       = "node-locations"
  cluster    = google_container_cluster.default_zonal.id
  node_count = 2

  node_locations = [
    "us-central1-a",
    "us-central1-b",
  ]
}

resource "google_container_node_pool" "initial_node_count_zonal" {
  name               = "initial-node-count"
  cluster            = google_container_cluster.default_zonal.id
  initial_node_count = 4
}

resource "google_container_node_pool" "autoscaling_zonal" {
  name    = "autoscaling"
  cluster = google_container_cluster.default_zonal.id

  autoscaling {
    min_node_count = 2
    max_node_count = 10
  }
}

resource "google_container_cluster" "zonal_usage" {
  name     = "default"
  location = "us-central1-a"
}

resource "google_container_cluster" "regional_usage" {
  name     = "default"
  location = "us-central1"
}

resource "google_container_cluster" "node_locations_usage" {
  name     = "node-locations"
  location = "us-central1"

  node_locations = [
    "us-central1-a",
    "us-central1-b",
  ]
}

resource "google_container_node_pool" "zonal_usage" {
  name       = "zonal"
  cluster    = google_container_cluster.zonal_usage.id
  node_count = 3
}

resource "google_container_node_pool" "regional_usage" {
  name       = "regional"
  cluster    = google_container_cluster.regional_usage.id
  node_count = 3
}

resource "google_container_node_pool" "node_locations_usage" {
  name       = "node-locations"
  cluster    = google_container_cluster.node_locations_usage.id
  node_count = 3
}
