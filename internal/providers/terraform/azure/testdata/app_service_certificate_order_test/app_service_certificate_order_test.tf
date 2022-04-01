provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_app_service_certificate_order" "standard_cert" {
  name                = "example-cert-order"
  resource_group_name = "fake"
  location            = "global"
  distinguished_name  = "CN=example.com"
}

resource "azurerm_app_service_certificate_order" "wildcard_cert" {
  name                = "example-cert-order"
  resource_group_name = "fake"
  location            = "global"
  distinguished_name  = "CN=example.com"
  product_type        = "WildCard"
}
