terraform {
  required_providers {
    azurerm = {
      source = "hashicorp/azurerm"
      version = "~>3.0"
    }
  }
}
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "rg" {
  name     = "appservice-rg"
  location = "westeurope"
}

resource "azurerm_app_service_plan" "appserviceplan" {
  name                = "appserviceplan"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  reserved = true
  kind = "Linux"

  sku {
    tier = "Standard"
    size ="S1"
  }
}

resource "azurerm_windows_web_app" "frontwebapp" {
  name                = "testazurermwebapp"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  service_plan_id     = azurerm_app_service_plan.appserviceplan.id

  site_config {}
}
