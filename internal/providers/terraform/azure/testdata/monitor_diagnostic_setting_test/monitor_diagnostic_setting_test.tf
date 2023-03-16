provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

data "azurerm_client_config" "current" {}

variable "fake_storage_id" {
  type    = string
  default = "/subscriptions/12345678-1234-5678-90ab-1234567890ab/resourceGroups/MyRG/providers/Microsoft.Storage/storageAccounts/MyST1"
}

variable "fake_log_analytics_id" {
  type    = string
  default = "/subscriptions/12345678-1234-5678-90ab-1234567890ab/resourceGroups/MyRG/providers/Microsoft.OperationalInsights/workspaces/MyWS1"
}

resource "azurerm_key_vault" "example" {
  name                = "examplekeyvault"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  tenant_id           = data.azurerm_client_config.current.tenant_id
  sku_name            = "standard"
}

resource "azurerm_monitor_diagnostic_setting" "example" {
  name               = "example"
  target_resource_id = azurerm_key_vault.example.id
  storage_account_id = var.fake_storage_id

  metric {
    category = "AllMetrics"

    retention_policy {
      enabled = false
    }
  }
}

resource "azurerm_monitor_diagnostic_setting" "example_with_usage" {
  name               = "example"
  target_resource_id = azurerm_key_vault.example.id
  storage_account_id = var.fake_storage_id

  metric {
    category = "AllMetrics"

    retention_policy {
      enabled = false
    }
  }
}

resource "azurerm_monitor_diagnostic_setting" "example_log_analytics_target" {
  name               = "example"
  target_resource_id = azurerm_key_vault.example.id

  log_analytics_workspace_id = var.fake_log_analytics_id

  metric {
    category = "AllMetrics"

    retention_policy {
      enabled = false
    }
  }
}
