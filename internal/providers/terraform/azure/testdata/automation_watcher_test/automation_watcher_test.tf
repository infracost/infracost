provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "eastus"
}

resource "azurerm_automation_account" "example" {
  name                = "example-account"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku_name            = "Basic"
}

resource "azurerm_automation_watcher" "example" {
  name                             = "example-watcher"
  location                         = azurerm_resource_group.example.location
  automation_account_id            = azurerm_automation_account.example.id
  script_name                      = "myscript.ps1"
  script_run_on                    = "HybridWorker"
  execution_frequency_in_seconds   = 300
}
