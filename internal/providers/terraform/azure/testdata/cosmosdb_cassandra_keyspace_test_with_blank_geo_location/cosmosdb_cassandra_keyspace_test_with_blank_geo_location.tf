provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "eastus"
}

resource "azurerm_cosmosdb_account" "with_blank_geo_location" {
  name                = "tfex-cosmosdb-account"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  offer_type          = "Standard"


  consistency_policy {
    consistency_level = "Strong"
  }

  backup {
    type = "Periodic"
  }
}

resource "azurerm_cosmosdb_cassandra_keyspace" "with_blank_geo_location" {
  name                = "tfex-cosmos-cassandra-keyspace"
  resource_group_name = azurerm_cosmosdb_account.with_blank_geo_location.resource_group_name
  account_name        = azurerm_cosmosdb_account.with_blank_geo_location.name
  autoscale_settings {
    max_throughput = 4000
  }
}
