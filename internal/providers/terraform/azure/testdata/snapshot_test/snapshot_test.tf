provider "azurerm" {
  features {}
  skip_provider_registration = true
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_managed_disk" "example" {
  name                 = "examplemd"
  location             = azurerm_resource_group.example.location
  resource_group_name  = azurerm_resource_group.example.name
  storage_account_type = "Standard_LRS"
  create_option        = "Empty"
  disk_size_gb         = 70
}

locals {
  permutations = [
    {
      name         = "managed"
      source_uri   = azurerm_managed_disk.example.id
      disk_size_gb = null
    },
    {
      name         = "source"
      source_uri   = null
      disk_size_gb = 20
    }
  ]
}

resource "azurerm_snapshot" "example" {
  for_each = { for entry in local.permutations : "${entry.name}" => entry }

  name                = each.value.name
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  create_option       = "Copy"
  source_uri          = each.value.source_uri
  disk_size_gb        = each.value.disk_size_gb
}

resource "azurerm_snapshot" "example_usage" {
  for_each = { for entry in local.permutations : "${entry.name}" => entry }

  name                = each.value.name
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  create_option       = "Copy"
  source_uri          = each.value.source_uri
  disk_size_gb        = each.value.disk_size_gb
}
