provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-notificationhub-resources"
  location = "eastus"
}
resource "azurerm_eventhub_namespace" "standard" {
  name                = "acceptanceTestEventHubNamespace"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "Standard"
  capacity            = 1

  tags = {
    environment = "Production"
  }
}

resource "azurerm_eventhub_namespace" "basic" {
  name                = "acceptanceTestEventHubNamespace"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "Basic"
  capacity            = 1

  tags = {
    environment = "Production"
  }
}

resource "azurerm_eventhub_namespace" "standardwithoutusage" {
  name                = "acceptanceTestEventHubNamespace"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "Standard"
  capacity            = 1

  tags = {
    environment = "Production"
  }
}

resource "azurerm_eventhub_namespace" "basicwithoutusage" {
  name                = "acceptanceTestEventHubNamespace"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "Basic"
  capacity            = 1

  tags = {
    environment = "Production"
  }
}