provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

locals {
  permutations = [
    {
      sku = "WS1"
    },
    {
      sku = "WS2"
    },
    {
      sku = "WS3"
    },
  ]
}

resource "azurerm_storage_account" "storage_account" {
  name                     = "icstorageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_app_service_plan" "app_service_plan" {
  for_each = { for entry in local.permutations : "${entry.sku}" => entry }

  name                = "app-service-plan"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  kind                = "elastic"

  sku {
    tier = "WorkflowStandard"
    size = each.value.sku
  }
}

resource "azurerm_logic_app_standard" "logic_app_standard" {
  for_each = { for entry in local.permutations : "${entry.sku}" => entry }

  name                       = "logic-app-standard"
  location                   = azurerm_resource_group.example.location
  resource_group_name        = azurerm_resource_group.example.name
  app_service_plan_id        = azurerm_app_service_plan.app_service_plan[each.key].id
  storage_account_name       = azurerm_storage_account.storage_account.name
  storage_account_access_key = azurerm_storage_account.storage_account.primary_access_key
}

resource "azurerm_logic_app_standard" "logic_app_standard_with_usage" {
  name                       = "logic-app-standard"
  location                   = azurerm_resource_group.example.location
  resource_group_name        = azurerm_resource_group.example.name
  app_service_plan_id        = azurerm_app_service_plan.app_service_plan["WS1"].id
  storage_account_name       = azurerm_storage_account.storage_account.name
  storage_account_access_key = azurerm_storage_account.storage_account.primary_access_key
}

resource "azurerm_logic_app_standard" "logic_app_standard_unknown_sku" {
  name                       = "logic-app-standard"
  location                   = azurerm_resource_group.example.location
  resource_group_name        = azurerm_resource_group.example.name
  app_service_plan_id        = "app-service-plan123"
  storage_account_name       = azurerm_storage_account.storage_account.name
  storage_account_access_key = azurerm_storage_account.storage_account.primary_access_key
}

resource "azurerm_logic_app_standard" "logic_app_standard_unknown_sku_with_usage" {
  name                       = "logic-app-standard"
  location                   = azurerm_resource_group.example.location
  resource_group_name        = azurerm_resource_group.example.name
  app_service_plan_id        = "app-service-plan123"
  storage_account_name       = azurerm_storage_account.storage_account.name
  storage_account_access_key = azurerm_storage_account.storage_account.primary_access_key
}

