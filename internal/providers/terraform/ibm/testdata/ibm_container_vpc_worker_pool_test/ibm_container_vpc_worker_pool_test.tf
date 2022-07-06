

terraform {
  required_providers {
    ibm = {
      source  = "IBM-Cloud/ibm"
      version = "~> 1.12.0"
    }
  }
}

provider "ibm" {
  region = "us-south"
}

resource "ibm_container_vpc_worker_pool" "test_vpc_worker_pool" {
  cluster          = "my_vpc_cluster"
  worker_pool_name = "my_vpc_pool"
  flavor           = "bx2.4x16"
  vpc_id           = "6015365a-9d93-4bb4-8248-79ae0db2dc21"
  worker_count     = "2"

  zones {
    name      = "us-south-1"
    subnet_id = "015ffb8b-efb1-4c03-8757-29335a07493b"
  }
}
