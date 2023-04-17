provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "exampleRG1"
  location = "eastus"
}

resource "azurerm_cdn_profile" "std_verizon" {
  name                = "example-cdn"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "Standard_Verizon"
}

resource "azurerm_cdn_profile" "prm_verizon" {
  name                = "example-cdn"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "Premium_Verizon"
}

resource "azurerm_cdn_profile" "std_microsoft" {
  name                = "example-cdn"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "Standard_Microsoft"
}

resource "azurerm_cdn_profile" "std_akamai" {
  name                = "example-cdn"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "Standard_Akamai"
}

resource "azurerm_cdn_endpoint" "std_verizon_with_opt" {
  name                = "example"
  profile_name        = azurerm_cdn_profile.std_verizon.name
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  optimization_type   = "DynamicSiteAcceleration"

  origin {
    name      = "example"
    host_name = "www.contoso.com"
  }
}

resource "azurerm_cdn_endpoint" "prm_verizon" {
  name                = "example"
  profile_name        = azurerm_cdn_profile.prm_verizon.name
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name

  origin {
    name      = "example"
    host_name = "www.contoso.com"
  }
}

resource "azurerm_cdn_endpoint" "std_microsoft" {
  name                = "example"
  profile_name        = azurerm_cdn_profile.std_microsoft.name
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  global_delivery_rule {
    cache_expiration_action {
      behavior = "SetIfMissing"
      duration = "00:05:00"
    }
  }
  delivery_rule {
    name  = "rule1"
    order = 1
  }
  delivery_rule {
    name  = "rule2"
    order = 2
  }
  delivery_rule {
    name  = "rule3"
    order = 3
  }
  delivery_rule {
    name  = "rule4"
    order = 4
  }
  delivery_rule {
    name  = "rule5"
    order = 5
  }
  delivery_rule {
    name  = "rule6"
    order = 6
  }
  delivery_rule {
    name  = "rule7"
    order = 7
  }

  origin {
    name      = "example"
    host_name = "www.contoso.com"
  }
}

resource "azurerm_cdn_endpoint" "std_akamai" {
  name                = "example"
  profile_name        = azurerm_cdn_profile.std_akamai.name
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name

  origin {
    name      = "example"
    host_name = "www.contoso.com"
  }
}

resource "azurerm_cdn_endpoint" "non_usage" {
  name                = "example"
  profile_name        = azurerm_cdn_profile.std_akamai.name
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  optimization_type   = "DynamicSiteAcceleration"

  origin {
    name      = "example"
    host_name = "www.contoso.com"
  }
}
