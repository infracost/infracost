provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "main" {
  name     = "example-resources"
  location = "westus"
}

resource "azurerm_virtual_network" "main" {
  name                = "network"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
}

resource "azurerm_subnet" "internal" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.main.name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = ["10.0.2.0/24"]
}

resource "azurerm_storage_account" "example" {
  name                     = "icstorageaccountname"
  resource_group_name      = azurerm_resource_group.main.name
  location                 = azurerm_resource_group.main.location
  account_kind             = "FileStorage"
  account_tier             = "Standard"
  account_replication_type = "LRS"
  access_tier              = "Cool"
}

resource "azurerm_private_endpoint" "without_usage_file" {
  name                = "example-privateendpoint"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  subnet_id           = azurerm_subnet.internal.id

  private_service_connection {
    name                           = "example-privateserviceconnection"
    private_connection_resource_id = azurerm_storage_account.example.id
    subresource_names              = ["file"]
    is_manual_connection           = false
  }
}

resource "azurerm_private_endpoint" "with_inbound" {
  name                = "example-privateendpoint"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  subnet_id           = azurerm_subnet.internal.id

  private_service_connection {
    name                           = "example-privateserviceconnection"
    private_connection_resource_id = azurerm_storage_account.example.id
    subresource_names              = ["file"]
    is_manual_connection           = false
  }
}

resource "azurerm_private_endpoint" "with_outbound" {
  name                = "example-privateendpoint"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  subnet_id           = azurerm_subnet.internal.id

  private_service_connection {
    name                           = "example-privateserviceconnection"
    private_connection_resource_id = azurerm_storage_account.example.id
    subresource_names              = ["file"]
    is_manual_connection           = false
  }
}

resource "azurerm_private_endpoint" "with_multiple_tiers" {
  name                = "example-privateendpoint"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  subnet_id           = azurerm_subnet.internal.id

  private_service_connection {
    name                           = "example-privateserviceconnection"
    private_connection_resource_id = azurerm_storage_account.example.id
    subresource_names              = ["file"]
    is_manual_connection           = false
  }
}


