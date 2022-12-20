module "database" {
  source = "./tf-rds"

  depends_on = [module.modulePost]
}

module "modulePost" {
  source = "./tf-module"
}
