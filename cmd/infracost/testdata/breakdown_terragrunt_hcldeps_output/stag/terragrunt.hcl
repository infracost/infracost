include {
  path = find_in_parent_folders()
}

terraform {
  source = "..//modules/example3"
}

inputs = {
  test = "test"
}
