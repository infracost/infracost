provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "nat-gateway-example-rg"
  location = "westus"
}

resource "azurerm_postgresql_flexible_server" "gp" {
  name                = "example-psqlflexibleserver"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  storage_mb          = 32768

  sku_name = "GP_Standard_D4s_v3"
}

resource "azurerm_postgresql_flexible_server" "mo" {
  name                = "example-psqlflexibleserver"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  storage_mb          = 65536

  sku_name = "MO_Standard_E4s_v3"
}

resource "azurerm_postgresql_flexible_server" "burstable" {
  name                = "example-psqlflexibleserver"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  storage_mb          = 131072

  sku_name = "B_Standard_B1ms"
}

resource "azurerm_postgresql_flexible_server" "non_usage_gp" {
  name                = "example-psqlflexibleserver"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location

  sku_name = "GP_Standard_D16s_v3"
}

resource "azurerm_postgresql_flexible_server" "readable_location_set" {
  name                = "readable-location"
  resource_group_name = "anything"
  location            = "East US"
  sku_name            = "B_Standard_B1ms"
}
