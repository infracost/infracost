provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_application_insights" "example" {
  name                = "tf-test-appinsights"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  application_type    = "web"
}

resource "azurerm_application_insights_standard_web_test" "example-disabled" {
  name                    = "example-test-default-frequency"
  resource_group_name     = azurerm_resource_group.example.name
  location                = "West Europe"
  application_insights_id = azurerm_application_insights.example.id
  geo_locations           = ["example"]

  enabled = false

  request {
    url = "http://www.example.com"
  }
}

resource "azurerm_application_insights_standard_web_test" "example-default-frequency" {
  name                    = "example-test-default-frequency"
  resource_group_name     = azurerm_resource_group.example.name
  location                = "West Europe"
  application_insights_id = azurerm_application_insights.example.id
  geo_locations           = ["example"]

  request {
    url = "http://www.example.com"
  }
}

resource "azurerm_application_insights_standard_web_test" "example" {
  for_each = toset(["300", "600", "900"])

  name                    = "example-test-frequency-${each.key}"
  resource_group_name     = azurerm_resource_group.example.name
  location                = "West Europe"
  application_insights_id = azurerm_application_insights.example.id
  geo_locations           = ["example"]

  frequency = each.key

  request {
    url = "http://www.example.com"
  }
}
