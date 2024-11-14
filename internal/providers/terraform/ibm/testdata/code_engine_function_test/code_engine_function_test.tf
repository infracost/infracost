resource "ibm_resource_group" "test_group" {
  name = "test-resource-group"
}

resource "ibm_code_engine_project" "ce_project" {
  name              = "ce_project"
  resource_group_id = ibm_resource_group.test_group.id
}

resource "ibm_code_engine_function" "ce_function" {
  project_id                    = ibm_code_engine_project.ce_project.id
  name                          = "ce_function"
  code_reference                = "icr.io/codeengine/helloworld"
  scale_memory_limit            = "4G"
  scale_cpu_limit               = "1"
  runtime                       = "nodejs-20"
}

resource "ibm_code_engine_function" "ce_function2" {
  project_id                    = ibm_code_engine_project.ce_project.id
  name                          = "ce_function2"
  code_reference                = "icr.io/codeengine/helloworld"
  scale_memory_limit            = "2G"
  scale_cpu_limit               = "0.5"
  runtime                       = "nodejs-20"
}
