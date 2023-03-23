provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_sql_server" "example" {
  name                         = "myexamplesqlserver"
  resource_group_name          = azurerm_resource_group.example.name
  location                     = "eastus"
  version                      = "12.0"
  administrator_login          = "admin"
  administrator_login_password = "password"
}

resource "azurerm_sql_elasticpool" "basic_100" {
  name                = "basic-100"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  server_name         = azurerm_sql_server.example.name
  edition             = "Basic"
  dtu                 = 100
  db_dtu_min          = 0
  db_dtu_max          = 5
  pool_size           = 5000
}

resource "azurerm_sql_elasticpool" "standard_200" {
  name                = "standard-200"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  server_name         = azurerm_sql_server.example.name
  edition             = "Standard"
  dtu                 = 200
  db_dtu_min          = 0
  db_dtu_max          = 5
  pool_size           = 307200
}

resource "azurerm_sql_elasticpool" "premium_500" {
  name                = "premium-500"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  server_name         = azurerm_sql_server.example.name
  edition             = "Premium"
  dtu                 = 500
  db_dtu_min          = 0
  db_dtu_max          = 5
  pool_size           = 1048576
}
