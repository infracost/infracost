provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
  region      = "us-central1"
}

resource "google_secret_manager_secret" "secret_example" {
  secret_id = "secret"

  labels = {
    label = "example"
  }

  replication {
    user_managed {
      replicas {
        location = "us-central1"
      }
      replicas {
        location = "us-east1"
      }
    }
  }
}

resource "google_secret_manager_secret" "secret_with_usage" {
  secret_id = "secret-with-usage"

  labels = {
    label = "example-with-usage"
  }

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret" "secret_replicas_with_usage" {
  secret_id = "secret-with-usage"

  labels = {
    label = "example-replicas-wth-usage"
  }

  replication {
    user_managed {
      replicas {
        location = "us-central1"
      }
      replicas {
        location = "us-east1"
      }
      replicas {
        location = "us-east2"
      }
    }
  }
}
