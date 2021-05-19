provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "eastus"
}

resource "azurerm_cosmosdb_account" "example" {
  name                = "tfex-cosmosdb-account"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  offer_type          = "Standard"
	analytical_storage_enabled = true

  capabilities {
    name = "EnableCassandra"
  }

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
    zone_redundant 		= true
  }

  geo_location {
    location          = "japanwest"
    failover_priority = 1
		zone_redundant 		= true
  }
}

resource "azurerm_cosmosdb_cassandra_keyspace" "autoscale" {
  name                = "tfex-cosmos-cassandra-keyspace"
  resource_group_name = azurerm_cosmosdb_account.example.resource_group_name
  account_name        = azurerm_cosmosdb_account.example.name
  autoscale_settings {
		max_throughput = 4000
	}
}

resource "azurerm_cosmosdb_cassandra_keyspace" "provisioned" {
  name                = "tfex-cosmos-cassandra-keyspace"
  resource_group_name = azurerm_cosmosdb_account.example.resource_group_name
  account_name        = azurerm_cosmosdb_account.example.name
	throughput = 500
}

resource "azurerm_cosmosdb_cassandra_keyspace" "serverless" {
  name                = "tfex-cosmos-cassandra-keyspace"
  resource_group_name = azurerm_cosmosdb_account.example.resource_group_name
  account_name        = azurerm_cosmosdb_account.example.name
}
