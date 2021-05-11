provider "azurerm" {
  skip_provider_registration = true
  features {}
  subscription_id             = "84171726-c002-4fc7-924a-b6ff82c59677"
  client_id                   = "2c517011-ceed-4166-9fc3-d1c6e077d0fa"
  client_secret               = "KhWvOLIR6Zkq4Yf9I_Ceg_jiOJRfCqlMHw"
  tenant_id                   = "b0b4f93e-59ae-44ff-b39a-e51e7e2b16e9"
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
  reserved = false

  sku {
    tier = "Standard"
    size = "EP2"
    capacity = 1 
  }
}
resource "azurerm_app_service_plan" "funcApp" {
  name                = "api-appserviceplan-pro"
  location            = azurerm_resource_group.example1.location
  resource_group_name = azurerm_resource_group.example1.name
  kind                = "FunctionApp"
  reserved = false

  sku {
    tier = "Standard"
    size = "EP2"
    capacity = 1 
  }
}

resource "azurerm_storage_account" "example" {
  name                     = "functionsapptestsa"
  resource_group_name      = azurerm_resource_group.example1.name
  location                 = azurerm_resource_group.example1.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
}
resource "azurerm_function_app" "my_functions" {
  name                       = "test-azure-functions"
  location                   = azurerm_resource_group.example1.location
  resource_group_name        = azurerm_resource_group.example1.name
  app_service_plan_id        = azurerm_app_service_plan.elastic.id
  storage_account_name       = azurerm_storage_account.example.name
  storage_account_access_key = azurerm_storage_account.example.primary_access_key
}
resource "azurerm_function_app" "my_functions1" {
  name                       = "test-azure-functions"
  location                   = azurerm_resource_group.example1.location
  resource_group_name        = azurerm_resource_group.example1.name
  app_service_plan_id        = azurerm_app_service_plan.funcApp.id
  storage_account_name       = azurerm_storage_account.example.name
  storage_account_access_key = azurerm_storage_account.example.primary_access_key
}