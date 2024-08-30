provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_storage_account" "example" {
  name                     = "icexamplestorageacc"
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

  identity {
    type = "SystemAssigned"
  }

  tags = {
    Env = "production"
  }
}

resource "azurerm_synapse_sql_pool" "default" {
  name                 = "examplesqlpool"
  synapse_workspace_id = azurerm_synapse_workspace.example.id
  sku_name             = "DW200c"
  create_mode          = "Default"
  storage_account_type = "GRS"
}

resource "azurerm_synapse_sql_pool" "storage" {
  name                 = "examplesqlpool"
  synapse_workspace_id = azurerm_synapse_workspace.example.id
  sku_name             = "DW200c"
  create_mode          = "Default"
  storage_account_type = "GRS"
}

resource "azurerm_synapse_sql_pool" "no_backup" {
  name                 = "examplesqlpool"
  synapse_workspace_id = azurerm_synapse_workspace.example.id
  sku_name             = "DW200c"
  create_mode          = "Default"
  storage_account_type = "GRS"
}
