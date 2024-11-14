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

resource "ibm_code_engine_app" "ce_app" {
  project_id                    = ibm_code_engine_project.ce_project.id
  name                          = "ce-app"
  image_reference               = "icr.io/codeengine/helloworld"
  scale_initial_instances       = 1
  scale_memory_limit            = "2G"
  scale_cpu_limit               = "1"
}

resource "ibm_code_engine_app" "ce_app2" {
  project_id                    = ibm_code_engine_project.ce_project.id
  name                          = "ce-app2"
  image_reference               = "icr.io/codeengine/helloworld"
}

resource "ibm_code_engine_app" "ce_app3" {
  project_id                    = ibm_code_engine_project.ce_project.id
  name                          = "ce-app3"
  image_reference               = "icr.io/codeengine/helloworld"
  scale_initial_instances       = 2
  scale_memory_limit            = "2G"
  scale_cpu_limit               = "1"
}

resource "ibm_code_engine_app" "ce_app4" {
  project_id                    = ibm_code_engine_project.ce_project.id
  name                          = "ce-app4"
  image_reference               = "icr.io/codeengine/helloworld"
  scale_initial_instances       = 3
  scale_memory_limit            = "2000M"
  scale_cpu_limit               = "1"
}