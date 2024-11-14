terraform {
  required_providers {
    ibm = {
      source = "IBM-Cloud/ibm"
    }
  }
}

provider "ibm" {
  region = "us-south"
}

resource "ibm_resource_group" "test_group" {
  name = "test-resource-group"
}

resource "ibm_code_engine_project" "ce_project" {
  name              = "ce_project"
  resource_group_id = ibm_resource_group.test_group.id
}

resource "ibm_code_engine_job" "ce_job" {
  project_id                    = ibm_code_engine_project.ce_project.id
  name                          = "ce-job"
  image_reference                = "icr.io/codeengine/helloworld"
  scale_memory_limit            = "4G"
  scale_cpu_limit               = "1"
}

resource "ibm_code_engine_job" "ce_job1" {
  project_id                    = ibm_code_engine_project.ce_project.id
  name                          = "ce-job1"
  image_reference                = "icr.io/codeengine/helloworld"
  scale_memory_limit            = "4G"
  scale_cpu_limit               = "1"
}
