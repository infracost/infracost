provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "deploy" {
  name     = "example-resources"
  location = "australiaeast"
}

resource "azurerm_virtual_network" "deploy" {
  name                = "deploy-vnet"
  location            = azurerm_resource_group.deploy.location
  resource_group_name = azurerm_resource_group.deploy.name
  address_space       = ["10.0.1.0/16"]
}

resource "azurerm_subnet" "deploy" {
  name                 = "deploy-subnet"
  resource_group_name  = azurerm_resource_group.deploy.name
  virtual_network_name = azurerm_virtual_network.deploy.name
  address_prefixes     = ["10.0.1.0/24"]
}

resource "azurerm_active_directory_domain_service" "standard" {
  name                = "example-aadds_1"
  location            = "australiaeast"
  resource_group_name = "aadds-rg"

  domain_name           = "widgetslogin.net"
  sku                   = "Standard"
  filtered_sync_enabled = false

  initial_replica_set {
    subnet_id = azurerm_subnet.deploy.id
  }

  tags = {
    Environment = "prod"
  }
}

resource "azurerm_active_directory_domain_service" "premium" {
  name                = "example-aadds_2"
  location            = "australiaeast"
  resource_group_name = "aadds-rg"

  domain_name           = "widgetslogin.net"
  sku                   = "Premium"
  filtered_sync_enabled = false

  initial_replica_set {
    subnet_id = azurerm_subnet.deploy.id
  }

  tags = {
    Environment = "prod"
  }
}

resource "azurerm_active_directory_domain_service" "enterprise" {
  name                = "example-aadds_3"
  location            = "australiaeast"
  resource_group_name = "aadds-rg"

  domain_name           = "widgetslogin.net"
  sku                   = "Enterprise"
  filtered_sync_enabled = false

  initial_replica_set {
    subnet_id = azurerm_subnet.deploy.id
  }

  tags = {
    Environment = "prod"
  }
}
