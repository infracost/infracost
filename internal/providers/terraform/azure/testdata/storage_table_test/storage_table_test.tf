provider "azurerm" {
  features {}
}

locals {
  account_replication_types = ["LRS", "ZRS", "GRS", "RA-GRS", "GZRS", "RA-GZRS"]
}

resource "azurerm_resource_group" "test" {
  name     = "test-resources"
  location = "West Europe"
}

resource "azurerm_storage_account" "standard" {
  for_each                 = toset(local.account_replication_types)
  name                     = "standard${each.value}"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  account_tier             = "Standard"
  account_kind             = "StorageV2"
  account_replication_type = each.value
}

resource "azurerm_storage_table" "standard" {
  for_each             = toset(local.account_replication_types)
  name                 = "tablestandard${each.value}"
  storage_account_name = azurerm_storage_account.standard[each.value].name
}

resource "azurerm_user_assigned_identity" "test" {
  name                = "test-identity"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
}

resource "azurerm_storage_account" "inline_key" {
  for_each                 = toset(local.account_replication_types)
  name                     = "customermanagedkey${each.value}"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  account_tier             = "Standard"
  account_kind             = "StorageV2"
  account_replication_type = each.value
  customer_managed_key {
    user_assigned_identity_id = azurerm_user_assigned_identity.test.id
  }
}

resource "azurerm_storage_table" "inline_key" {
  for_each             = toset(local.account_replication_types)
  name                 = "tableinlinekey${each.value}"
  storage_account_name = azurerm_storage_account.inline_key[each.value].name
}

resource "azurerm_key_vault" "test" {
  name                = "testkeyvault"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  tenant_id           = "00000000-0000-0000-0000-000000000000"
  sku_name            = "standard"
}

resource "azurerm_storage_account" "external_key" {
  name                     = "externalkey"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  account_tier             = "Standard"
  account_kind             = "StorageV2"
  account_replication_type = "LRS"
}

resource "azurerm_storage_account_customer_managed_key" "external_key" {
  storage_account_id = azurerm_storage_account.external_key.id
  key_name           = "keyname"
  key_vault_id       = azurerm_key_vault.test.id
}

resource "azurerm_storage_table" "external_key" {
  name                 = "tableexternalkey"
  storage_account_name = azurerm_storage_account.external_key.name
}

resource "azurerm_storage_table" "usage" {
  name                 = "tableusage"
  storage_account_name = azurerm_storage_account.standard["LRS"].name
}