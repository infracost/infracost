provider "azurerm" {
  skip_provider_registration = true
  features {}
}
resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_dns_zone" "example" {
  name                = "mydomain.com"
  resource_group_name = azurerm_resource_group.example.name
}

resource "azurerm_dns_aaaa_record" "over1B" {
  name                = "test"
  zone_name           = azurerm_dns_zone.example.name
  resource_group_name = azurerm_resource_group.example.name
  ttl                 = 300
  records             = ["2001:db8::1:0:0:1"]
}
resource "azurerm_dns_aaaa_record" "first1B" {
  name                = "test"
  zone_name           = azurerm_dns_zone.example.name
  resource_group_name = azurerm_resource_group.example.name
  ttl                 = 300
  records             = ["2001:db8::1:0:0:1"]
}
resource "azurerm_dns_aaaa_record" "withoutUsage" {
  name                = "test"
  zone_name           = azurerm_dns_zone.example.name
  resource_group_name = azurerm_resource_group.example.name
  ttl                 = 300
  records             = ["2001:db8::1:0:0:1"]
}
