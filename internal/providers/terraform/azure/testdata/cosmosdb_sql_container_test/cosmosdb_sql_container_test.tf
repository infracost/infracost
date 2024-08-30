provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "eastus"
}

resource "azurerm_cosmosdb_account" "example" {
  name                       = "tfex-cosmosdb-account"
  resource_group_name        = azurerm_resource_group.example.name
  location                   = azurerm_resource_group.example.location
  offer_type                 = "Standard"
  analytical_storage_enabled = true

  consistency_policy {
    consistency_level = "Strong"
  }

  geo_location {
    location          = "westus"
    failover_priority = 0
  }

  backup {
    type                = "Periodic"
    interval_in_minutes = 240
    retention_in_hours  = 16
  }
}

resource "azurerm_cosmosdb_account" "continuous_backup" {
  name                = "tfex-cosmosdb-account"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  offer_type          = "Standard"

  consistency_policy {
    consistency_level = "Strong"
  }

  geo_location {
    location          = "westus"
    failover_priority = 0
  }

  geo_location {
    location          = "centralus"
    failover_priority = 1
    zone_redundant    = true
  }
  backup {
    type = "Continuous"
  }
}

resource "azurerm_cosmosdb_account" "multi-master_backup2copies" {
  name                = "tfex-cosmosdb-account"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  offer_type          = "Standard"


  consistency_policy {
    consistency_level = "Strong"
  }

  geo_location {
    location          = "westus"
    failover_priority = 0
  }

  geo_location {
    location          = "centralus"
    failover_priority = 1
  }
  backup {
    type = "Periodic"
  }
}

resource "azurerm_cosmosdb_sql_database" "example" {
  name                = "tfex-cosmos-mongo-db"
  resource_group_name = azurerm_cosmosdb_account.example.resource_group_name
  account_name        = azurerm_cosmosdb_account.example.name
  autoscale_settings {
    max_throughput = 4000
  }
}

resource "azurerm_cosmosdb_sql_container" "non-usage_autoscale" {
  name                = "example-container"
  resource_group_name = azurerm_cosmosdb_account.example.resource_group_name
  account_name        = azurerm_cosmosdb_account.example.name
  database_name       = azurerm_cosmosdb_sql_database.example.name
  partition_key_paths = ["/definition/id"]
  autoscale_settings {
    max_throughput = 4000
  }
}

resource "azurerm_cosmosdb_sql_container" "autoscale" {
  name                = "example-container"
  resource_group_name = azurerm_cosmosdb_account.example.resource_group_name
  account_name        = azurerm_cosmosdb_account.continuous_backup.name
  database_name       = azurerm_cosmosdb_sql_database.example.name
  partition_key_paths = ["/definition/id"]
  autoscale_settings {
    max_throughput = 6000
  }
}

resource "azurerm_cosmosdb_sql_container" "provisioned" {
  name                = "example-container"
  resource_group_name = azurerm_cosmosdb_account.example.resource_group_name
  account_name        = azurerm_cosmosdb_account.continuous_backup.name
  database_name       = azurerm_cosmosdb_sql_database.example.name
  partition_key_paths = ["/definition/id"]
  throughput          = 500
}

resource "azurerm_cosmosdb_sql_container" "mutli-master_backup2copies" {
  name                = "example-container"
  resource_group_name = azurerm_cosmosdb_account.example.resource_group_name
  account_name        = azurerm_cosmosdb_account.example.name
  database_name       = azurerm_cosmosdb_sql_database.example.name
  partition_key_paths = ["/definition/id"]
  throughput          = 1000
}

resource "azurerm_cosmosdb_sql_container" "serverless" {
  name                = "example-container"
  resource_group_name = azurerm_cosmosdb_account.example.resource_group_name
  account_name        = azurerm_cosmosdb_account.example.name
  database_name       = azurerm_cosmosdb_sql_database.example.name
  partition_key_paths = ["/definition/id"]
}
