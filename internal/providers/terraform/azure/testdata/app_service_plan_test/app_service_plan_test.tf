provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "exampleRG1"
  location = "eastus"
}

resource "azurerm_app_service_plan" "standard_s1" {
  name                = "api-appserviceplan-pro"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  kind                = "Windows"
  reserved            = false

  sku {
    tier     = "Standard"
    size     = "S1"
    capacity = 1
  }
}

resource "azurerm_app_service_plan" "standard_s2" {
  name                = "api-appserviceplan-pro"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  kind                = "Windows"
  reserved            = false

  sku {
    tier     = "Standard"
    size     = "S1"
    capacity = 5
  }
}

resource "azurerm_app_service_plan" "premium_v1" {
  name                = "api-appserviceplan-pro"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  kind                = "Linux"
  reserved            = false

  sku {
    tier     = "PremiumContainer"
    size     = "P1v1"
    capacity = 1
  }
}


resource "azurerm_app_service_plan" "isolated" {
  name                = "api-appserviceplan-pro"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  kind                = "Linux"
  reserved            = false

  sku {
    tier     = "Isolated"
    size     = "I1"
    capacity = 1
  }
}


resource "azurerm_app_service_plan" "isolated_v2" {
  name                = "api-appserviceplan-pro"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  kind                = "Linux"
  reserved            = false

  sku {
    tier     = "Isolated"
    size     = "I1v2"
    capacity = 1
  }
}

resource "azurerm_app_service_plan" "premium_v2" {
  name                = "api-appserviceplan-pro"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  kind                = "Linux"
  reserved            = false

  sku {
    tier     = "PremiumContainer"
    size     = "P1v2"
    capacity = 10
  }
}

resource "azurerm_app_service_plan" "basic" {
  name                = "api-appserviceplan-pro"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  kind                = "Linux"
  reserved            = false

  sku {
    tier     = "Basic"
    size     = "B2"
    capacity = 1
  }
}

resource "azurerm_app_service_plan" "pc2" {
  name                = "api-appserviceplan-pro"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  kind                = "Windows"
  reserved            = false

  sku {
    tier     = "PremiumContainer"
    size     = "PC2"
    capacity = 1
  }
}

resource "azurerm_app_service_plan" "pc3" {
  name                = "api-appserviceplan-pro"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  kind                = "Windows"
  reserved            = false

  sku {
    tier     = "PremiumContainer"
    size     = "PC3"
    capacity = 15
  }
}

resource "azurerm_app_service_plan" "default_capacity" {
  name                = "api-appserviceplan-pro"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  kind                = "Linux"
  reserved            = false

  sku {
    tier = "Basic"
    size = "B2"
  }
}
