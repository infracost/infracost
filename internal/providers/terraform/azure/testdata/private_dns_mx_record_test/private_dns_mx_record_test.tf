provider "azurerm" {
  skip_provider_registration = true
  features {}
}
resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_private_dns_zone" "example" {
  name                = "mydomain.com"
  resource_group_name = azurerm_resource_group.example.name
}

resource "azurerm_private_dns_mx_record" "over1B" {
  name                = "example"
  resource_group_name = azurerm_resource_group.example.name
  zone_name           = azurerm_private_dns_zone.example.name
  ttl                 = 300

  record {
    preference = 10
    exchange   = "mx1.contoso.com"
  }

  record {
    preference = 20
    exchange   = "backupmx.contoso.com"
  }

  tags = {
    Environment = "Production"
  }
}
resource "azurerm_private_dns_mx_record" "first1B" {
  name                = "example"
  resource_group_name = azurerm_resource_group.example.name
  zone_name           = azurerm_private_dns_zone.example.name
  ttl                 = 300

  record {
    preference = 10
    exchange   = "mx1.contoso.com"
  }

  record {
    preference = 20
    exchange   = "backupmx.contoso.com"
  }

  tags = {
    Environment = "Production"
  }
}
resource "azurerm_private_dns_mx_record" "withoutUsage" {
  name                = "example"
  resource_group_name = azurerm_resource_group.example.name
  zone_name           = azurerm_private_dns_zone.example.name
  ttl                 = 300

  record {
    preference = 10
    exchange   = "mx1.contoso.com"
  }

  record {
    preference = 20
    exchange   = "backupmx.contoso.com"
  }

  tags = {
    Environment = "Production"
  }
}
