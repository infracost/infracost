provider "azurerm" {
  skip_provider_registration = true
  features {}
}
resource "azurerm_resource_group" "test" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_private_dns_zone" "test" {
  name                = "mydomain.com"
  resource_group_name = azurerm_resource_group.test.name
}
resource "azurerm_private_dns_txt_record" "over1B" {
  name                = "test"
  resource_group_name = azurerm_resource_group.test.name
  zone_name           = azurerm_private_dns_zone.test.name
  ttl                 = 300

  record {
    value = "v=spf1 mx ~all"
  }
}
resource "azurerm_private_dns_txt_record" "first1B" {
  name                = "test"
  resource_group_name = azurerm_resource_group.test.name
  zone_name           = azurerm_private_dns_zone.test.name
  ttl                 = 300

  record {
    value = "v=spf1 mx ~all"
  }
}
resource "azurerm_private_dns_txt_record" "withoutUsage" {
  name                = "test"
  resource_group_name = azurerm_resource_group.test.name
  zone_name           = azurerm_private_dns_zone.test.name
  ttl                 = 300

  record {
    value = "v=spf1 mx ~all"
  }
}
