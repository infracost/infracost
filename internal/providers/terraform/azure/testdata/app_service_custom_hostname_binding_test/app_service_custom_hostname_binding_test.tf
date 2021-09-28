provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "random_id" "server" {
  keepers = {
    azi_id = 1
  }

  byte_length = 8
}

resource "azurerm_resource_group" "example" {
  name     = "some-resource-group"
  location = "eastus2"
}

resource "azurerm_app_service_plan" "example" {
  name                = "some-app-service-plan"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name

  sku {
    tier     = "Standard"
    size     = "S1"
    capacity = 1
  }
}

resource "azurerm_app_service" "example" {
  name                = random_id.server.hex
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  app_service_plan_id = azurerm_app_service_plan.example.id
}

resource "azurerm_app_service_custom_hostname_binding" "example" {
  hostname            = "www.mywebsite.com"
  app_service_name    = azurerm_app_service.example.name
  resource_group_name = azurerm_resource_group.example.name
  ssl_state           = "IpBasedEnabled"
}

resource "azurerm_app_service_custom_hostname_binding" "example1" {
  hostname            = "www.mywebsite.com"
  app_service_name    = azurerm_app_service.example.name
  resource_group_name = azurerm_resource_group.example.name
  ssl_state           = "SniEnabled"
}

resource "azurerm_app_service_custom_hostname_binding" "no_ssl_state" {
  hostname            = "www.mywebsite.com"
  app_service_name    = azurerm_app_service.example.name
  resource_group_name = azurerm_resource_group.example.name
}
