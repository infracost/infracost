provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

module "ecs_service" {
  source      = "./_ecs"
  task_cpu    = module.configs.result["ECS_TASK_CPU"]
  task_memory = module.configs.result["ECS_TASK_RAM"]
}

module "configs" {
  source    = "./_config"
  base_path = "./config/"
  files = [
    "config.base.json"
  ]
}

module "ecs_service_with_path_module" {
  source      = "./_ecs"
  task_cpu    = module.configs_with_path_module.result["ECS_TASK_CPU"]
  task_memory = module.configs_with_path_module.result["ECS_TASK_RAM"]
}

module "configs_with_path_module" {
  source = "./_config_with_path_module"
}

module "ecs_service_with_path_cwd" {
  source      = "./_ecs"
  task_cpu    = module.configs_with_path_cwd.result["ECS_TASK_CPU"]
  task_memory = module.configs_with_path_cwd.result["ECS_TASK_RAM"]
}

module "configs_with_path_cwd" {
  source = "./_config_with_path_cwd"
}
