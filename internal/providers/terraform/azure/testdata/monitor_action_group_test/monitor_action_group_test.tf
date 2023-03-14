provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "eastus"
}

resource "azurerm_monitor_action_group" "example" {
  name                = "CriticalAlertsAction"
  resource_group_name = azurerm_resource_group.example.name
  short_name          = "p0action"
}

resource "azurerm_monitor_action_group" "example_with_usage" {
  name                = "CriticalAlertsActionWithUsage"
  resource_group_name = azurerm_resource_group.example.name
  short_name          = "p0action"
}

resource "azurerm_resource_group" "global_example" {
  name     = "example-resources"
  location = "global"
}

resource "azurerm_monitor_action_group" "global_example_with_usage" {
  name                = "CriticalAlertsActionWithUsage"
  resource_group_name = azurerm_resource_group.global_example.name
  short_name          = "p0action"
}
