provider "azurerm" {
	skip_provider_registration = true
  features {}
}

resource "azurerm_app_service_certificate_binding" "ip_ssl" {
  hostname_binding_id = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Web/sites/mywebappfake/hostNameBindings/example.example.com"
  certificate_id      = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Web/certificates/example.example.com"
  ssl_state           = "IpBasedEnabled"
}

resource "azurerm_app_service_certificate_binding" "sni_ssl" {
  hostname_binding_id = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Web/sites/mywebappfake/hostNameBindings/example.example.com"
  certificate_id      = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Web/certificates/example.example.com"
  ssl_state           = "SniEnabled"
}
