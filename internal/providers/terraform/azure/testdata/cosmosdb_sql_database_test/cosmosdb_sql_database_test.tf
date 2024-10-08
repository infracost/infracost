provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "eastus"
}

resource "azurerm_cosmosdb_account" "example" {
  name                       = "tfex-cosmosdb-account-example"
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
  name                = "tfex-cosmosdb-account-continuous-backup"
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
  name                = "tfex-cosmosdb-account-backup2copies"
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

resource "azurerm_cosmosdb_sql_database" "non-usage_autoscale" {
  name                = "tfex-cosmos-mongo-db"
  resource_group_name = azurerm_cosmosdb_account.example.resource_group_name
  account_name        = azurerm_cosmosdb_account.example.name
  autoscale_settings {
    max_throughput = 4000
  }
}

resource "azurerm_cosmosdb_sql_database" "autoscale" {
  name                = "tfex-cosmos-mongo-db"
  resource_group_name = azurerm_cosmosdb_account.example.resource_group_name
  account_name        = azurerm_cosmosdb_account.continuous_backup.name
  autoscale_settings {
    max_throughput = 6000
  }
}

resource "azurerm_cosmosdb_sql_database" "provisioned" {
  name                = "tfex-cosmos-mongo-db"
  resource_group_name = azurerm_cosmosdb_account.example.resource_group_name
  account_name        = azurerm_cosmosdb_account.continuous_backup.name
  throughput          = 500
}

resource "azurerm_cosmosdb_sql_database" "mutli-master_backup2copies" {
  name                = "tfex-cosmos-mongo-db"
  resource_group_name = azurerm_cosmosdb_account.example.resource_group_name
  account_name        = azurerm_cosmosdb_account.multi-master_backup2copies.name
  throughput          = 1000
}

resource "azurerm_cosmosdb_sql_database" "serverless" {
  name                = "tfex-cosmos-mongo-db"
  resource_group_name = azurerm_cosmosdb_account.example.resource_group_name
  account_name        = azurerm_cosmosdb_account.example.name
}

resource "azurerm_cosmosdb_sql_database" "account_by_name" {
  name                = "tfex-cosmos-account-by-name"
  resource_group_name = azurerm_cosmosdb_account.example.resource_group_name
  account_name        = "tfex-cosmosdb-account-example"
  throughput          = 500
}
