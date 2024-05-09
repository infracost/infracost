provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_user_assigned_identity" "example" {
  location            = azurerm_resource_group.example.location
  name                = "example"
  resource_group_name = azurerm_resource_group.example.name
}

resource "azurerm_federated_identity_credential" "base" {
  name                = "example"
  resource_group_name = azurerm_resource_group.example.name
  audience            = ["foo"]
  issuer              = "https://foo"
  parent_id           = azurerm_user_assigned_identity.example.id
  subject             = "foo"
}

resource "azurerm_federated_identity_credential" "usage_p1" {
  name                = "example"
  resource_group_name = azurerm_resource_group.example.name
  audience            = ["foo"]
  issuer              = "https://foo"
  parent_id           = azurerm_user_assigned_identity.example.id
  subject             = "foo"
}

resource "azurerm_federated_identity_credential" "usage_p2" {
  name                = "example"
  resource_group_name = azurerm_resource_group.example.name
  audience            = ["foo"]
  issuer              = "https://foo"
  parent_id           = azurerm_user_assigned_identity.example.id
  subject             = "foo"
}
