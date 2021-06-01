provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-notificationhub-resources"
  location = "eastus"
}

resource "azurerm_notification_hub_namespace" "basicbelow10M" {
  name                = "myappnamespace"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  namespace_type      = "NotificationHub"

  sku_name = "Basic"
}

resource "azurerm_notification_hub_namespace" "basicabove10M" {
  name                = "myappnamespace"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  namespace_type      = "NotificationHub"

  sku_name = "Basic"
}

resource "azurerm_notification_hub_namespace" "basicabove100M" {
  name                = "myappnamespace"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  namespace_type      = "NotificationHub"

  sku_name = "Basic"
}

resource "azurerm_notification_hub_namespace" "stdbelow10M" {
  name                = "myappnamespace"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  namespace_type      = "NotificationHub"

  sku_name = "Standard"
}

resource "azurerm_notification_hub_namespace" "stdabove10M" {
  name                = "myappnamespace"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  namespace_type      = "NotificationHub"

  sku_name = "Standard"
}

resource "azurerm_notification_hub_namespace" "basicwithoutUsage" {
  name                = "myappnamespace"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  namespace_type      = "NotificationHub"

  sku_name = "Basic"
}

resource "azurerm_notification_hub_namespace" "standardwithoutUsage" {
  name                = "myappnamespace"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  namespace_type      = "NotificationHub"

  sku_name = "Standard"
}

resource "azurerm_notification_hub_namespace" "stdabove100M" {
  name                = "myappnamespace"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  namespace_type      = "NotificationHub"

  sku_name = "Standard"
}

resource "azurerm_notification_hub_namespace" "free" {
  name                = "myappnamespace"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  namespace_type      = "NotificationHub"

  sku_name = "Free"
}

