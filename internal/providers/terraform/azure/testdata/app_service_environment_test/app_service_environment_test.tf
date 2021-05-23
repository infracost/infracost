provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "exampleRG1"
  location = "eastus"
}

resource "azurerm_subnet" "ase" {
  name                 = "asesubnet"
  resource_group_name  = azurerm_resource_group.example.name
  virtual_network_name = azurerm_virtual_network.example.name
  address_prefixes     = ["10.0.1.0/24"]
}

resource "azurerm_virtual_network" "example" {
  name                = "example-vnet1"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  address_space       = ["10.0.0.0/16"]
}

resource "azurerm_app_service_environment" "example_I1" {
  name                         = "example-ase"
  subnet_id                    = azurerm_subnet.ase.id
  pricing_tier                 = "I1"
  front_end_scale_factor       = 10
  internal_load_balancing_mode = "Web, Publishing"
  allowed_user_ip_cidrs        = ["11.22.33.44/32", "55.66.77.0/24"]
  resource_group_name          = azurerm_resource_group.example.name


  cluster_setting {
    name  = "DisableTls1.0"
    value = "1"
  }
}

resource "azurerm_app_service_environment" "example_I2" {
  name                         = "example-ase"
  subnet_id                    = azurerm_subnet.ase.id
  pricing_tier                 = "I2"
  front_end_scale_factor       = 10
  internal_load_balancing_mode = "Web, Publishing"
  allowed_user_ip_cidrs        = ["11.22.33.44/32", "55.66.77.0/24"]
  resource_group_name          = azurerm_resource_group.example.name


  cluster_setting {
    name  = "DisableTls1.0"
    value = "1"
  }
}


resource "azurerm_app_service_environment" "linux_I1" {
  name                         = "example-ase"
  subnet_id                    = azurerm_subnet.ase.id
  pricing_tier                 = "I1"
  front_end_scale_factor       = 10
  internal_load_balancing_mode = "Web, Publishing"
  allowed_user_ip_cidrs        = ["11.22.33.44/32", "55.66.77.0/24"]
  resource_group_name          = azurerm_resource_group.example.name


  cluster_setting {
    name  = "DisableTls1.0"
    value = "1"
  }
}

resource "azurerm_app_service_environment" "linux_I2" {
  name                         = "example-ase"
  subnet_id                    = azurerm_subnet.ase.id
  pricing_tier                 = "I2"
  front_end_scale_factor       = 10
  internal_load_balancing_mode = "Web, Publishing"
  allowed_user_ip_cidrs        = ["11.22.33.44/32", "55.66.77.0/24"]
  resource_group_name          = azurerm_resource_group.example.name


  cluster_setting {
    name  = "DisableTls1.0"
    value = "1"
  }
}

resource "azurerm_app_service_environment" "windows_I1" {
  name                         = "example-ase"
  subnet_id                    = azurerm_subnet.ase.id
  pricing_tier                 = "I1"
  front_end_scale_factor       = 10
  internal_load_balancing_mode = "Web, Publishing"
  allowed_user_ip_cidrs        = ["11.22.33.44/32", "55.66.77.0/24"]
  resource_group_name          = azurerm_resource_group.example.name


  cluster_setting {
    name  = "DisableTls1.0"
    value = "1"
  }
}

resource "azurerm_app_service_environment" "windows_I2" {
  name                         = "example-ase"
  subnet_id                    = azurerm_subnet.ase.id
  pricing_tier                 = "I2"
  front_end_scale_factor       = 10
  internal_load_balancing_mode = "Web, Publishing"
  allowed_user_ip_cidrs        = ["11.22.33.44/32", "55.66.77.0/24"]
  resource_group_name          = azurerm_resource_group.example.name


  cluster_setting {
    name  = "DisableTls1.0"
    value = "1"
  }
}
