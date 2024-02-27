provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "eastus"
}

resource "azurerm_redis_cache" "standard_c3" {
  name                = "example-cache"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  capacity            = 3
  family              = "C"
  sku_name            = "Standard"
}

resource "azurerm_redis_cache" "basic_c2" {
  name                = "example-cache"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  capacity            = 4
  family              = "C"
  sku_name            = "Basic"
}

resource "azurerm_redis_cache" "premium_p1_australiacentral" {
  name                = "example-cache"
  location            = "australiacentral"
  resource_group_name = azurerm_resource_group.example.name
  capacity            = 1
  shard_count         = 1
  family              = "P"
  sku_name            = "Premium"
}

resource "azurerm_redis_cache" "premium_p1" {
  name                = "example-cache"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  capacity            = 1
  family              = "P"
  sku_name            = "Premium"
  shard_count         = 3
}

resource "azurerm_redis_cache" "premium_p2_replicas_per_master" {
  name                = "example-cache"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  capacity            = 2
  family              = "P"
  sku_name            = "Premium"
  shard_count         = 3
  replicas_per_master = 2
}

resource "azurerm_redis_cache" "premium_p3_replicas_per_primary" {
  name                 = "example-cache"
  location             = azurerm_resource_group.example.location
  resource_group_name  = azurerm_resource_group.example.name
  capacity             = 3
  family               = "P"
  sku_name             = "Premium"
  shard_count          = 3
  replicas_per_primary = 3
}

resource "azurerm_redis_cache" "premium_zero_shards" {
  name                = "example-cache"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  capacity            = 1
  family              = "P"
  sku_name            = "Premium"
  shard_count         = 0
}


