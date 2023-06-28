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
resource "azurerm_private_dns_srv_record" "over1B" {
  name                = "test"
  resource_group_name = azurerm_resource_group.test.name
  zone_name           = azurerm_private_dns_zone.test.name
  ttl                 = 300

  record {
    priority = 1
    weight   = 5
    port     = 8080
    target   = "target1.contoso.com"
  }

  record {
    priority = 10
    weight   = 10
    port     = 8080
    target   = "target2.contoso.com"
  }

  tags = {
    Environment = "Production"
  }
}
resource "azurerm_private_dns_srv_record" "first1B" {
  name                = "test"
  resource_group_name = azurerm_resource_group.test.name
  zone_name           = azurerm_private_dns_zone.test.name
  ttl                 = 300

  record {
    priority = 1
    weight   = 5
    port     = 8080
    target   = "target1.contoso.com"
  }

  record {
    priority = 10
    weight   = 10
    port     = 8080
    target   = "target2.contoso.com"
  }

  tags = {
    Environment = "Production"
  }
}
resource "azurerm_private_dns_srv_record" "withoutUsage" {
  name                = "test"
  resource_group_name = azurerm_resource_group.test.name
  zone_name           = azurerm_private_dns_zone.test.name
  ttl                 = 300

  record {
    priority = 1
    weight   = 5
    port     = 8080
    target   = "target1.contoso.com"
  }

  record {
    priority = 10
    weight   = 10
    port     = 8080
    target   = "target2.contoso.com"
  }

  tags = {
    Environment = "Production"
  }
}
