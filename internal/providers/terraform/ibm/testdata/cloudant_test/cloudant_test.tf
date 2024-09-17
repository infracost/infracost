
terraform {
  required_providers {
    ibm = {
      source  = "IBM-Cloud/ibm"
      version = "1.63.0"
    }
  }
}
provider "ibm" {
  region = "us-south"
}

resource "ibm_cloudant" "standard_cloudant" {
  name     = "standard-cloudant"
  location = "us-south"
  plan     = "standard"

  legacy_credentials  = true
  include_data_events = false
  capacity            = 1
  enable_cors         = true

  cors_config {
    allow_credentials = false
    origins           = ["https://example.com"]
  }

  timeouts {
    create = "15m"
    update = "15m"
    delete = "15m"
  }
}

resource "ibm_cloudant" "standard_cloudant_without_usage" {
  name     = "standard-cloudant-without-usage"
  location = "us-south"
  plan     = "standard"

  legacy_credentials  = true
  include_data_events = false
  capacity            = 1
  enable_cors         = true

  cors_config {
    allow_credentials = false
    origins           = ["https://example.com"]
  }

  timeouts {
    create = "15m"
    update = "15m"
    delete = "15m"
  }
}

resource "ibm_cloudant" "standard_exceeded_storage" {
  name     = "standard-exceeded-storage"
  location = "us-south"
  plan     = "standard"

  legacy_credentials  = true
  include_data_events = false
  capacity            = 1
  enable_cors         = true

  cors_config {
    allow_credentials = false
    origins           = ["https://example.com"]
  }

  timeouts {
    create = "15m"
    update = "15m"
    delete = "15m"
  }
}

resource "ibm_cloudant" "lite_cloudant" {
  name     = "lite-cloudant"
  location = "us-south"
  plan     = "lite"

  legacy_credentials  = true
  include_data_events = false
  capacity            = 1
  enable_cors         = true

  cors_config {
    allow_credentials = false
    origins           = ["https://example.com"]
  }

  timeouts {
    create = "15m"
    update = "15m"
    delete = "15m"
  }
}

resource "ibm_cloudant" "dedicated_cloudant" {
  name     = "dedicated-cloudant"
  location = "us-south"
  plan     = "dedicated-hardware"

  legacy_credentials  = true
  include_data_events = false
  capacity            = 1
  enable_cors         = true

  cors_config {
    allow_credentials = false
    origins           = ["https://example.com"]
  }

  timeouts {
    create = "15m"
    update = "15m"
    delete = "15m"
  }
}
