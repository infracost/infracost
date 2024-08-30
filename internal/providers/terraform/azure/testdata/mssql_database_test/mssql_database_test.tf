provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "eastus"
}

resource "azurerm_mssql_server" "example" {
  name                         = "example-sqlserver"
  resource_group_name          = azurerm_resource_group.example.name
  location                     = "eastus"
  version                      = "12.0"
  administrator_login          = "fake"
  administrator_login_password = "fake"
}

resource "azurerm_mssql_database" "general_purpose_gen_without_license" {
  name         = "acctest-db-d"
  server_id    = azurerm_mssql_server.example.id
  sku_name     = "GP_Gen5_4"
  license_type = "BasePrice"
}
resource "azurerm_mssql_database" "business_critical_gen" {
  name        = "acctest-db-d"
  server_id   = azurerm_mssql_server.example.id
  sku_name    = "BC_Gen5_8"
  max_size_gb = 10
}
resource "azurerm_mssql_database" "business_critical_m" {
  name        = "acctest-db-d"
  server_id   = azurerm_mssql_server.example.id
  sku_name    = "BC_M_8"
  max_size_gb = 50
}
resource "azurerm_mssql_database" "hyperscale_gen" {
  name        = "acctest-db-d"
  server_id   = azurerm_mssql_server.example.id
  sku_name    = "HS_Gen5_2"
  max_size_gb = 100
}

resource "azurerm_mssql_database" "hyperscale_gen_with_replicas" {
  name               = "acctest-db-d"
  server_id          = azurerm_mssql_server.example.id
  sku_name           = "HS_Gen5_2"
  read_replica_count = 2
}

resource "azurerm_mssql_database" "general_purpose_gen" {
  name         = "acctest-db-d"
  server_id    = azurerm_mssql_server.example.id
  sku_name     = "GP_Gen5_4"
  license_type = "LicenseIncluded"
}

resource "azurerm_mssql_database" "general_purpose_gen_zone" {
  name           = "acctest-db-d"
  server_id      = azurerm_mssql_server.example.id
  sku_name       = "GP_Gen5_4"
  zone_redundant = true
}

resource "azurerm_mssql_database" "serverless" {
  name      = "acctest-db-d"
  server_id = azurerm_mssql_server.example.id
  sku_name  = "GP_S_Gen5_4"
}

resource "azurerm_mssql_database" "serverless_zone" {
  name           = "acctest-db-d"
  server_id      = azurerm_mssql_server.example.id
  sku_name       = "GP_S_Gen5_4"
  zone_redundant = true
}

resource "azurerm_mssql_database" "backup_default" {
  name      = "acctest-db-d"
  server_id = azurerm_mssql_server.example.id
  sku_name  = "GP_Gen5_4"
}

resource "azurerm_mssql_database" "backup_geo" {
  name                 = "acctest-db-d"
  server_id            = azurerm_mssql_server.example.id
  sku_name             = "GP_Gen5_4"
  storage_account_type = "Geo"
}

resource "azurerm_mssql_database" "backup_zone" {
  name                 = "acctest-db-d"
  server_id            = azurerm_mssql_server.example.id
  sku_name             = "GP_Gen5_4"
  storage_account_type = "Zone"
}

resource "azurerm_mssql_database" "backup_local" {
  name                 = "acctest-db-d"
  server_id            = azurerm_mssql_server.example.id
  sku_name             = "GP_Gen5_4"
  storage_account_type = "Local"
}

resource "azurerm_mssql_database" "standard1" {
  name      = "acctest-db-d"
  server_id = azurerm_mssql_server.example.id
  sku_name  = "S1"
}

resource "azurerm_mssql_database" "standard12" {
  name        = "acctest-db-d"
  server_id   = azurerm_mssql_server.example.id
  sku_name    = "S12"
  max_size_gb = 500
}

resource "azurerm_mssql_database" "premium6" {
  name      = "acctest-db-d"
  server_id = azurerm_mssql_server.example.id
  sku_name  = "P6"
}

resource "azurerm_mssql_database" "blank_sku" {
  name      = "acctest-db-e"
  server_id = azurerm_mssql_server.example.id
}

resource "azurerm_mssql_database" "elastic_pool" {
  name      = "acctest-db-f"
  server_id = azurerm_mssql_server.example.id
  sku_name  = "ElasticPool"
}
