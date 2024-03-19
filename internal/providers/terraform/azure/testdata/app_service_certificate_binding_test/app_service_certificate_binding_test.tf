provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "webapp"
  location = "westus2"
}

resource "azurerm_app_service_certificate" "example" {
  name                = "example-cert"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  password            = "terraform"
  pfx_blob            = "someblob"
}

resource "azurerm_app_service_certificate_binding" "example" {
  hostname_binding_id = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Web/sites/mywebappfake/hostNameBindings/example.example.com"
  certificate_id      = azurerm_app_service_certificate.example.id
  ssl_state           = "IpBasedEnabled"
}
resource "azurerm_app_service_certificate_binding" "withoutId" {
  hostname_binding_id = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Web/sites/mywebappfake/hostNameBindings/example.example.com"
  certificate_id      = azurerm_app_service_certificate.example.id
  ssl_state           = "SniEnabled"
}

