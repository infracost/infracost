provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_storage_account" "Premium_ZRS" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = "eastus"
  account_kind             = "BlockBlobStorage"
  account_tier             = "Premium"
  account_replication_type = "ZRS"
}

resource "azurerm_storage_account" "Premium_LRS" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = "eastus"
  account_kind             = "BlockBlobStorage"
  account_tier             = "Premium"
  account_replication_type = "LRS"
}

resource "azurerm_storage_account" "Standard_LRS_Hot" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = "eastus"
  account_kind             = "BlockBlobStorage"
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_storage_account" "Standard_LRS_Cool" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = "eastus"
  account_kind             = "BlockBlobStorage"
  account_tier             = "Standard"
  account_replication_type = "LRS"
  access_tier              = "Cool"
}

resource "azurerm_storage_account" "Standard_GRS_Hot" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = "eastus"
  account_kind             = "BlockBlobStorage"
  account_tier             = "Standard"
  account_replication_type = "GRS"
}

resource "azurerm_storage_account" "Standard_GRS_Cool" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = "eastus"
  account_kind             = "BlockBlobStorage"
  account_tier             = "Standard"
  account_replication_type = "GRS"
  access_tier              = "Cool"
}

resource "azurerm_storage_account" "Standard_RAGRS_Hot" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = "eastus"
  account_kind             = "BlockBlobStorage"
  account_tier             = "Standard"
  account_replication_type = "RAGRS"
}

resource "azurerm_storage_account" "Standard_RAGRS_Cool" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = "eastus"
  account_kind             = "BlockBlobStorage"
  account_tier             = "Standard"
  account_replication_type = "RAGRS"
  access_tier              = "Cool"
}

resource "azurerm_storage_account" "without_usage_file" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = "eastus"
  account_kind             = "BlockBlobStorage"
  account_tier             = "Standard"
  account_replication_type = "RAGRS"
  access_tier              = "Cool"
}