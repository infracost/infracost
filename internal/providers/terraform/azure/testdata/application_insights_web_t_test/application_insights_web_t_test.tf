provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "exampleRG1"
  location = "eastus"
}

resource "azurerm_application_insights" "example" {
  name                = "tf-test-appinsights"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  application_type    = "web"
}

resource "azurerm_application_insights_web_test" "free" {
  name                    = "tf-test-appinsights-webtest"
  location                = azurerm_application_insights.example.location
  resource_group_name     = azurerm_resource_group.example.name
  application_insights_id = azurerm_application_insights.example.id
  kind                    = "ping"
  geo_locations           = ["us-tx-sn1-azr", "us-il-ch1-azr"]

  configuration = <<XML
  XML
}

resource "azurerm_application_insights_web_test" "non_free" {
  name                    = "tf-test-appinsights-webtest"
  location                = azurerm_application_insights.example.location
  resource_group_name     = azurerm_resource_group.example.name
  application_insights_id = azurerm_application_insights.example.id
  geo_locations           = ["us-tx-sn1-azr", "us-il-ch1-azr"]
  kind                    = "multistep"
  enabled                 = true


  configuration = <<XML
  XML
}
