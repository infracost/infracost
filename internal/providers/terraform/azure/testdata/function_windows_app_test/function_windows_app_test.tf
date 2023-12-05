provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example1" {
  name     = "exampleRG1"
  location = "eastus"
}

resource "azurerm_app_service_plan" "legacy_elastic" {
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

resource "azurerm_storage_account" "example" {
  name                     = "icfunctionsapptestsa"
  resource_group_name      = azurerm_resource_group.example1.name
  location                 = azurerm_resource_group.example1.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
  access_tier              = "Cool"
}

locals {
  os_types = ["Windows"]
  skus     = ["EP1", "EP2", "EP3", "Y1"]

  permutations = distinct(flatten([
    for os_type in local.os_types : [
      for sku in local.skus : {
        sku     = sku
        os_type = os_type
      }
    ]
  ]))

  tiers = ["Standard", "ElasticPremium"]
  kinds = ["elastic", "FunctionApp"]
  legacy_permutations = distinct(flatten([
    for tier in local.tiers : [
      for kind in local.kinds : [
        for sku in local.skus : {
          sku  = sku
          tier = tier
          kind = kind
        }
      ]
    ]
  ]))
}

resource "azurerm_service_plan" "plan" {
  for_each = { for entry in local.permutations : "${entry.os_type}.${entry.sku}" => entry }

  name                = "plan-${each.value.os_type}-${each.value.sku}"
  location            = azurerm_resource_group.example1.location
  resource_group_name = azurerm_resource_group.example1.name
  sku_name            = each.value.sku
  os_type             = each.value.os_type
}

resource "azurerm_app_service_plan" "legacy_plan" {
  for_each = { for entry in local.legacy_permutations : "${entry.tier}.${entry.sku}.${entry.kind}" => entry }

  name                = "legacy-plan-${each.value.tier}-${each.value.sku}-${each.value.kind}"
  location            = azurerm_resource_group.example1.location
  resource_group_name = azurerm_resource_group.example1.name
  kind                = each.value.kind
  reserved            = false

  sku {
    tier     = each.value.tier
    size     = each.value.sku
    capacity = 1
  }
}


resource "azurerm_windows_function_app" "function" {
  for_each = { for entry in azurerm_service_plan.plan : "${entry.name}" => entry }

  name                       = each.value.name
  location                   = azurerm_resource_group.example1.location
  resource_group_name        = azurerm_resource_group.example1.name
  service_plan_id            = each.value.id
  storage_account_name       = azurerm_storage_account.example.name
  storage_account_access_key = azurerm_storage_account.example.primary_access_key

  site_config {}
}

resource "azurerm_windows_function_app" "function_with_usage" {
  for_each = { for entry in azurerm_service_plan.plan : "${entry.name}" => entry }

  name                       = each.value.name
  location                   = azurerm_resource_group.example1.location
  resource_group_name        = azurerm_resource_group.example1.name
  service_plan_id            = each.value.id
  storage_account_name       = azurerm_storage_account.example.name
  storage_account_access_key = azurerm_storage_account.example.primary_access_key

  site_config {}
}

resource "azurerm_windows_function_app" "legacy_service_plan_function" {
  for_each = { for entry in azurerm_app_service_plan.legacy_plan : "${entry.name}" => entry }

  name                       = each.value.name
  location                   = azurerm_resource_group.example1.location
  resource_group_name        = azurerm_resource_group.example1.name
  service_plan_id            = each.value.id
  storage_account_name       = azurerm_storage_account.example.name
  storage_account_access_key = azurerm_storage_account.example.primary_access_key

  site_config {}
}

resource "azurerm_windows_function_app" "legacy_service_plan_function_with_usage" {
  for_each = { for entry in azurerm_app_service_plan.legacy_plan : "${entry.name}" => entry }

  name                       = each.value.name
  location                   = azurerm_resource_group.example1.location
  resource_group_name        = azurerm_resource_group.example1.name
  service_plan_id            = each.value.id
  storage_account_name       = azurerm_storage_account.example.name
  storage_account_access_key = azurerm_storage_account.example.primary_access_key

  site_config {}
}
