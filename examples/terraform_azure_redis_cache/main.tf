provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "main" {
  name     = "example-resources"
  location = var.region
}

resource "azurerm_virtual_network" "main" {
  name                = "deploy-vnet"
  location            = var.region
  resource_group_name = azurerm_resource_group.main.name
  address_space       = ["10.0.1.0/16"]
}

resource "azurerm_subnet" "main" {
  name                 = "deploy-subnet"
  resource_group_name  = azurerm_resource_group.main.name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = ["10.0.1.0/24"]
}

resource "azurerm_redis_cache" "main" {
  name                = var.name
  location            = var.region
  resource_group_name = azurerm_resource_group.main.name

  redis_version        = var.redis.redis_version
  replicas_per_primary = var.redis.replicas_per_primary
  family               = "P"
  capacity             = var.redis.capacity
  sku_name             = "Premium"
  subnet_id            = azurerm_subnet.main.id

  identity {
    type = "SystemAssigned"
  }

  public_network_access_enabled = false

  shard_count = var.cluster.shard_count

  tags = var.tags
}
