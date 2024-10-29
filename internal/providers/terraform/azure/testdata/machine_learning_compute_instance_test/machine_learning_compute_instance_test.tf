provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

locals {
  vm_sizes = [
    "Standard_D2s_v3",
    "Standard_D4s_v3",
    "Standard_D8s_v3",
    "Standard_D16s_v3",
    "Standard_DS12_v2",
    "Standard_F4s_v2",
    "Standard_F8s_v2",
    "Standard_F16s_v2",
    "Standard_E4s_v3",
    "Standard_E8s_v3",
    "Standard_E16s_v3",
    "Standard_NC6",
    "Standard_NC12",
    "Standard_NC24",
    "Standard_NV6",
    "Standard_NV12",
    "Standard_NV24",
  ]
}

resource "azurerm_storage_account" "example" {
  name                     = "examplestorageaccount"
  location                 = azurerm_resource_group.example.location
  resource_group_name      = azurerm_resource_group.example.name
  account_tier             = "Standard"
  account_replication_type = "GRS"
}

resource "azurerm_machine_learning_workspace" "example" {
  name                    = "example-workspace"
  location                = azurerm_resource_group.example.location
  resource_group_name     = azurerm_resource_group.example.name
  key_vault_id            = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group/providers/Microsoft.KeyVault/vaults/example"
  application_insights_id = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group/providers/Microsoft.Insights/components/example"

  identity {
    type = "SystemAssigned"
  }

  storage_account_id = azurerm_storage_account.example.id
}

resource "azurerm_machine_learning_compute_instance" "example" {
  for_each = toset(local.vm_sizes)

  name                          = "example-instance"
  machine_learning_workspace_id = azurerm_machine_learning_workspace.example.id
  virtual_machine_size          = each.value

}

resource "azurerm_machine_learning_compute_instance" "with_monthly_hrs" {
  name                          = "with-monthly-hrs"
  machine_learning_workspace_id = azurerm_machine_learning_workspace.example.id
  virtual_machine_size          = "Standard_D2s_v3"
}
