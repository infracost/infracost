provider "azurerm" {
  skip_provider_registration = true
  features {}
}

data "azurerm_resource_group" "main" {
  name = "test"
}

resource "azurerm_mssql_server" "blank_location" {
  name                         = "blank-location"
  resource_group_name          = data.azurerm_resource_group.main.name
  location                     = data.azurerm_resource_group.main.location
  version                      = "12.0"
  administrator_login          = "fake"
  administrator_login_password = "fake"
}

resource "azurerm_mssql_database" "blank_server_location" {
  name      = "acctest-db-e"
  sku_name  = "S3"
  server_id = azurerm_mssql_server.blank_location.id
}
