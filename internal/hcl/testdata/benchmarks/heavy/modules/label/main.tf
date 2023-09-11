module "base" {
  source = "cloudposse/label/null"

  namespace = "eg"
  stage     = "prod"
  name      = "bastion"
  delimiter = "-"

  tags = {
    "BusinessUnit" = "XYZ",
    "Snapshot"     = "true"
  }
}

module "child" {
  source = "cloudposse/label/null"

  attributes = ["abc"]

  tags = {
    "BusinessUnit" = "ABC"
  }

  context = module.base.context
}

output "context" {
  value = module.child.context
}
