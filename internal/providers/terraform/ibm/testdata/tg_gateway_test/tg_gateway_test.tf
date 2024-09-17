terraform {
  required_providers {
    ibm = {
      source  = "IBM-Cloud/ibm"
      version = "1.63.0"
    }
  }
}

provider "ibm" {
  generation = 2
  region     = "us-south"
}

resource "ibm_tg_gateway" "new_tg_gw" {
  name           = "transit-gateway-1"
  location       = "us-south"
  global         = true
  resource_group = "30951d2dff914dafb26455a88c0c0092"
}
