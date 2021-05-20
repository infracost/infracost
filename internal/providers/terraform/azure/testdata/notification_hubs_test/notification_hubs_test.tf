provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-notificationhub-resources"
  location = "eastus"
}

resource "azurerm_notification_hub_namespace" "example" {
  name                = "myappnamespace"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  namespace_type      = "NotificationHub"

  sku_name = "Basic"
}

resource "azurerm_notification_hub" "example" {
  name                = "mynotificationhub"
  namespace_name      = azurerm_notification_hub_namespace.example.name
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
}
