provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "eastus"
}

resource "azurerm_managed_redis" "balanced" {
  name                = "balanced-managed-redis"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  sku_name            = "Balanced_B3"

  default_database {}
}

resource "azurerm_managed_redis" "balanced_no_high_availability" {
  name                      = "balanced-no-ha-managed-redis"
  resource_group_name       = azurerm_resource_group.example.name
  location                  = azurerm_resource_group.example.location
  sku_name                  = "Balanced_B3"
  high_availability_enabled = false

  default_database {}
}

resource "azurerm_managed_redis" "memory_optimized" {
  name                = "memory-managed-redis"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  sku_name            = "MemoryOptimized_M10"

  default_database {}
}

resource "azurerm_managed_redis" "compute_optimized" {
  name                = "compute-managed-redis"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  sku_name            = "ComputeOptimized_X10"

  default_database {}
}

resource "azurerm_managed_redis" "flash_optimized" {
  name                = "flash-managed-redis"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  sku_name            = "FlashOptimized_A700"

  default_database {}
}
