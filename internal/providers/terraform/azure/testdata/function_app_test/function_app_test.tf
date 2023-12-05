provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example1" {
  name     = "exampleRG1"
  location = "eastus"
}
resource "azurerm_app_service_plan" "elastic" {
  name                = "api-appserviceplan-pro"
  location            = azurerm_resource_group.example1.location
  resource_group_name = azurerm_resource_group.example1.name
  kind                = "elastic"
  reserved            = false

  sku {
    tier     = "Standard"
    size     = "EP2"
    capacity = 1
  }
}

resource "azurerm_app_service_plan" "elasticPremium" {
  name                = "api-appserviceplan-elasticPremium"
  location            = azurerm_resource_group.example1.location
  resource_group_name = azurerm_resource_group.example1.name
  kind                = "elastic"
  reserved            = false

  sku {
    tier     = "ElasticPremium"
    size     = "EP1"
    capacity = 1
  }
}

resource "azurerm_app_service_plan" "funcApp" {
  name                = "api-appserviceplan-pro"
  location            = azurerm_resource_group.example1.location
  resource_group_name = azurerm_resource_group.example1.name
  kind                = "FunctionApp"
  reserved            = false

  sku {
    tier     = "Standard"
    size     = "ep2"
    capacity = 1
  }
}

resource "azurerm_storage_account" "example" {
  name                     = "icfunctionsapptestsa"
  resource_group_name      = azurerm_resource_group.example1.name
  location                 = azurerm_resource_group.example1.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
  access_tier              = "Cool"
}

resource "azurerm_function_app" "elasticFunction" {
  name                       = "elasticFunction"
  location                   = azurerm_resource_group.example1.location
  resource_group_name        = azurerm_resource_group.example1.name
  app_service_plan_id        = azurerm_app_service_plan.elastic.id
  storage_account_name       = azurerm_storage_account.example.name
  storage_account_access_key = azurerm_storage_account.example.primary_access_key
}
resource "azurerm_function_app" "elasticFunctionWithUsage" {
  name                       = "elasticFunctionWithUsage"
  location                   = azurerm_resource_group.example1.location
  resource_group_name        = azurerm_resource_group.example1.name
  app_service_plan_id        = azurerm_app_service_plan.elastic.id
  storage_account_name       = azurerm_storage_account.example.name
  storage_account_access_key = azurerm_storage_account.example.primary_access_key
}
resource "azurerm_function_app" "elasticFunctionWithZeroInstances" {
  name                       = "elasticFunctionWithZeroInstances"
  location                   = azurerm_resource_group.example1.location
  resource_group_name        = azurerm_resource_group.example1.name
  app_service_plan_id        = azurerm_app_service_plan.elastic.id
  storage_account_name       = azurerm_storage_account.example.name
  storage_account_access_key = azurerm_storage_account.example.primary_access_key
}
resource "azurerm_function_app" "elasticPremiumFunction" {
  name                       = "elasticPremiumFunction"
  location                   = azurerm_resource_group.example1.location
  resource_group_name        = azurerm_resource_group.example1.name
  app_service_plan_id        = azurerm_app_service_plan.elasticPremium.id
  storage_account_name       = azurerm_storage_account.example.name
  storage_account_access_key = azurerm_storage_account.example.primary_access_key
}

resource "azurerm_function_app" "functionApp" {
  name                       = "functionApp"
  location                   = azurerm_resource_group.example1.location
  resource_group_name        = azurerm_resource_group.example1.name
  app_service_plan_id        = azurerm_app_service_plan.funcApp.id
  storage_account_name       = azurerm_storage_account.example.name
  storage_account_access_key = azurerm_storage_account.example.primary_access_key
}

resource "azurerm_function_app" "functionAppWithAllUsage" {
  name                       = "functionAppWithAllUsage"
  location                   = azurerm_resource_group.example1.location
  resource_group_name        = azurerm_resource_group.example1.name
  app_service_plan_id        = azurerm_app_service_plan.funcApp.id
  storage_account_name       = azurerm_storage_account.example.name
  storage_account_access_key = azurerm_storage_account.example.primary_access_key
}

resource "azurerm_function_app" "functionAppWithLessThanMins" {
  name                       = "functionAppWithLessThanMins"
  location                   = azurerm_resource_group.example1.location
  resource_group_name        = azurerm_resource_group.example1.name
  app_service_plan_id        = azurerm_app_service_plan.funcApp.id
  storage_account_name       = azurerm_storage_account.example.name
  storage_account_access_key = azurerm_storage_account.example.primary_access_key
}

resource "azurerm_function_app" "functionAppWithOnlyExecutions" {
  name                       = "functionAppWithOnlyExecutions"
  location                   = azurerm_resource_group.example1.location
  resource_group_name        = azurerm_resource_group.example1.name
  app_service_plan_id        = azurerm_app_service_plan.funcApp.id
  storage_account_name       = azurerm_storage_account.example.name
  storage_account_access_key = azurerm_storage_account.example.primary_access_key
}

resource "azurerm_function_app" "functionAppWithMissingExecutions" {
  name                       = "functionAppWithMissingExecutions"
  location                   = azurerm_resource_group.example1.location
  resource_group_name        = azurerm_resource_group.example1.name
  app_service_plan_id        = azurerm_app_service_plan.funcApp.id
  storage_account_name       = azurerm_storage_account.example.name
  storage_account_access_key = azurerm_storage_account.example.primary_access_key
}

resource "azurerm_function_app" "functionAppNoAvailableServicePlan" {
  name                       = "functionAppNoAvailableServicePlan"
  location                   = azurerm_resource_group.example1.location
  resource_group_name        = azurerm_resource_group.example1.name
  app_service_plan_id        = "in_another_module"
  storage_account_name       = azurerm_storage_account.example.name
  storage_account_access_key = azurerm_storage_account.example.primary_access_key
}

resource "azurerm_function_app" "functionAppNoAvailableServicePlanButHasUsage" {
  name                       = "functionAppNoAvailableServicePlanButHasUsage"
  location                   = azurerm_resource_group.example1.location
  resource_group_name        = azurerm_resource_group.example1.name
  app_service_plan_id        = "in_another_module"
  storage_account_name       = azurerm_storage_account.example.name
  storage_account_access_key = azurerm_storage_account.example.primary_access_key
}
