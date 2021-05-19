provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "eastus"
}

resource "azurerm_mysql_server" "basic_2core" {
  name                = "example-mariadb-server"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name

  administrator_login          = "fake"
  administrator_login_password = "fake"

  sku_name   = "B_Gen5_2"
  storage_mb = 5120
  version    = "5.7"

  geo_redundant_backup_enabled = false
  ssl_enforcement_enabled      = true
}

resource "azurerm_mysql_server" "gp_4core" {
  name                = "example-mariadb-server"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name

  administrator_login          = "fake"
  administrator_login_password = "fake"

  sku_name   = "GP_Gen5_4"
  storage_mb = 4096000
  version    = "5.7"

  geo_redundant_backup_enabled = false
  ssl_enforcement_enabled      = true
}

resource "azurerm_mysql_server" "mo_16core" {
  name                = "example-mariadb-server"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name

  administrator_login          = "fake"
  administrator_login_password = "fake"

  sku_name   = "MO_Gen5_16"
  storage_mb = 5120
  version    = "5.7"

  geo_redundant_backup_enabled = false
  ssl_enforcement_enabled      = true
}

resource "azurerm_mysql_server" "without_geo" {
  name                = "example-mariadb-server"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name

  administrator_login          = "fake"
  administrator_login_password = "fake"

  sku_name   = "MO_Gen5_16"
  storage_mb = 5120
  version    = "5.7"

  geo_redundant_backup_enabled = false
  ssl_enforcement_enabled      = true
}

resource "azurerm_mysql_server" "with_geo" {
  name                = "example-mariadb-server"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name

  administrator_login          = "fake"
  administrator_login_password = "fake"

  sku_name   = "GP_Gen5_4"
  storage_mb = 4096000
  version    = "5.7"

  geo_redundant_backup_enabled = true
  ssl_enforcement_enabled      = true
}
