provider "azurerm" {
  features {}
  skip_provider_registration = true
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_kubernetes_cluster" "free_D2V2" {
  name                = "example-aks1"
  location            = "eastus"
  resource_group_name = azurerm_resource_group.example.name
  dns_prefix          = "exampleaks1"


  default_node_pool {
    name    = "default"
    vm_size = "Standard_D2_v2"
  }

  identity {
    type = "SystemAssigned"
  }
}

resource "azurerm_kubernetes_cluster" "paid_D2SV2_3nc_128gb" {
  name                = "example-aks1"
  location            = "eastus"
  resource_group_name = azurerm_resource_group.example.name
  dns_prefix          = "exampleaks1"
  sku_tier            = "Paid"

  default_node_pool {
    name            = "default"
    node_count      = 3
    vm_size         = "Standard_DS2_v2"
    os_disk_size_gb = 128
  }

  identity {
    type = "SystemAssigned"
  }
}

resource "azurerm_kubernetes_cluster" "paid_5nc_32gb" {
  name                = "example-aks1"
  location            = "eastus"
  resource_group_name = azurerm_resource_group.example.name
  dns_prefix          = "exampleaks1"
  sku_tier            = "Paid"

  default_node_pool {
    name            = "default"
    node_count      = 5
    vm_size         = "Standard_D2_v2"
    os_disk_size_gb = 32
  }

  identity {
    type = "SystemAssigned"
  }
}

resource "azurerm_kubernetes_cluster" "usage_ephemeral" {
  name                = "example-aks1"
  location            = "eastus"
  resource_group_name = azurerm_resource_group.example.name
  dns_prefix          = "exampleaks1"
  sku_tier            = "Paid"
  network_profile {
    network_plugin    = "azure"
    load_balancer_sku = "Standard"
  }
  addon_profile {
    http_application_routing {
      enabled = true
    }
  }
  default_node_pool {
    name         = "default"
    vm_size      = "Standard_D2_v2"
    os_disk_type = "Ephemeral"
  }

  identity {
    type = "SystemAssigned"
  }
}

