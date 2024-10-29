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

resource "azurerm_sql_database" "default_sql_database" {
  name                = "myexamplesqldatabase1"
  resource_group_name = azurerm_resource_group.example.name
  location            = "eastus"
  server_name         = azurerm_sql_server.example.name
}

resource "azurerm_sql_database" "sql_database_with_max_size" {
  name                = "myexamplesqldatabase2"
  resource_group_name = azurerm_resource_group.example.name
  location            = "eastus"
  server_name         = azurerm_sql_server.example.name
  max_size_bytes      = 10737418240
}

resource "azurerm_sql_database" "sql_database_with_edition_critical" {
  name                = "myexamplesqldatabase3"
  resource_group_name = azurerm_resource_group.example.name
  location            = "eastus"
  server_name         = azurerm_sql_server.example.name
  edition             = "BusinessCritical"
}

resource "azurerm_sql_database" "sql_database_with_edition_critical_zone_redundant" {
  name                = "myexamplesqldatabase3"
  resource_group_name = azurerm_resource_group.example.name
  location            = "eastus"
  server_name         = azurerm_sql_server.example.name
  edition             = "BusinessCritical"
  zone_redundant      = true
}

resource "azurerm_sql_database" "sql_database_with_edition_gen" {
  name                = "myexamplesqldatabase4"
  resource_group_name = azurerm_resource_group.example.name
  location            = "eastus"
  server_name         = azurerm_sql_server.example.name
  edition             = "GeneralPurpose"
}

resource "azurerm_sql_database" "sql_database_with_edition_gen_zone_redundant" {
  name                = "myexamplesqldatabase4"
  resource_group_name = azurerm_resource_group.example.name
  location            = "eastus"
  server_name         = azurerm_sql_server.example.name
  edition             = "GeneralPurpose"
  zone_redundant      = true
}

resource "azurerm_sql_database" "sql_database_with_edition_hyper" {
  name                = "myexamplesqldatabase5"
  resource_group_name = azurerm_resource_group.example.name
  location            = "eastus"
  server_name         = azurerm_sql_server.example.name
  edition             = "Hyperscale"
}

resource "azurerm_sql_database" "sql_database_with_edition_standard" {
  name                = "myexamplesqldatabase5"
  resource_group_name = azurerm_resource_group.example.name
  location            = "eastus"
  server_name         = azurerm_sql_server.example.name
  edition             = "Standard"
}

resource "azurerm_sql_database" "sql_database_with_edition_premium" {
  name                = "myexamplesqldatabase5"
  resource_group_name = azurerm_resource_group.example.name
  location            = "eastus"
  server_name         = azurerm_sql_server.example.name
  edition             = "Premium"
}

resource "azurerm_sql_database" "sql_database_with_service_object_name" {
  name                             = "myexamplesqldatabase9"
  resource_group_name              = azurerm_resource_group.example.name
  location                         = "eastus"
  server_name                      = azurerm_sql_server.example.name
  requested_service_objective_name = "BC_Gen5_2"
}

resource "azurerm_sql_database" "serverless" {
  name                             = "myexamplesqldatabase9"
  resource_group_name              = azurerm_resource_group.example.name
  location                         = "eastus"
  server_name                      = azurerm_sql_server.example.name
  requested_service_objective_name = "GP_S_Gen5_4"
}

resource "azurerm_sql_database" "backup" {
  name                             = "myexamplesqldatabase9"
  resource_group_name              = azurerm_resource_group.example.name
  location                         = "eastus"
  server_name                      = azurerm_sql_server.example.name
  requested_service_objective_name = "GP_Gen5_4"
}

resource "azurerm_sql_database" "premium6" {
  name                             = "myexamplesqldatabase9"
  resource_group_name              = azurerm_resource_group.example.name
  location                         = "eastus"
  server_name                      = azurerm_sql_server.example.name
  requested_service_objective_name = "P6"
}
