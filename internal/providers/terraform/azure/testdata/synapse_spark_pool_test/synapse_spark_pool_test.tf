provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_storage_account" "example" {
  name                     = "examplestorageacc"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
  account_kind             = "StorageV2"
  access_tier              = "Cool"
  is_hns_enabled           = "true"
}

resource "azurerm_storage_data_lake_gen2_filesystem" "example" {
  name               = "example"
  storage_account_id = azurerm_storage_account.example.id
}

resource "azurerm_synapse_workspace" "example" {
  name                                 = "example"
  resource_group_name                  = azurerm_resource_group.example.name
  location                             = azurerm_resource_group.example.location
  storage_data_lake_gen2_filesystem_id = azurerm_storage_data_lake_gen2_filesystem.example.id
  sql_administrator_login              = "sqladminuser"
  sql_administrator_login_password     = "H@Sh1CoR3!"
  managed_virtual_network_enabled      = false

  aad_admin {
    login     = "AzureAD Admin"
    object_id = "00000000-0000-0000-0000-000000000000"
    tenant_id = "00000000-0000-0000-0000-000000000000"
  }

  identity {
    type = "SystemAssigned"
  }

  tags = {
    Env = "production"
  }
}

resource "azurerm_synapse_spark_pool" "default" {
  name                 = "example"
  synapse_workspace_id = azurerm_synapse_workspace.example.id
  node_size_family     = "MemoryOptimized"
  node_size            = "Small"

  node_count = 3
  auto_pause {
    delay_in_minutes = 15
  }
}

resource "azurerm_synapse_spark_pool" "autoscale" {
  name                 = "example"
  synapse_workspace_id = azurerm_synapse_workspace.example.id
  node_size_family     = "MemoryOptimized"
  node_size            = "Small"

  auto_scale {
    min_node_count = 4
    max_node_count = 8
  }

  auto_pause {
    delay_in_minutes = 15
  }
}
