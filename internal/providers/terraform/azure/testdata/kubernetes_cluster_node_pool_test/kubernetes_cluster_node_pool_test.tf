provider "azurerm" {
  features {}
  skip_provider_registration = true
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_kubernetes_cluster" "example" {
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

resource "azurerm_kubernetes_cluster_node_pool" "example" {
  name                  = "internal"
  kubernetes_cluster_id = azurerm_kubernetes_cluster.example.id
  vm_size               = "Standard_DS2_v2"
}

resource "azurerm_kubernetes_cluster_node_pool" "basic_A2" {
  name                  = "internal"
  kubernetes_cluster_id = azurerm_kubernetes_cluster.example.id
  vm_size               = "Basic_A2"
}

resource "azurerm_kubernetes_cluster_node_pool" "Standard_DS2_v2" {
  name                  = "internal"
  kubernetes_cluster_id = azurerm_kubernetes_cluster.example.id
  vm_size               = "Standard_DS2_v2"
  node_count            = 2
  os_disk_type          = "Ephemeral"
}

resource "azurerm_kubernetes_cluster_node_pool" "usage_basic_A2" {
  name                  = "internal"
  kubernetes_cluster_id = azurerm_kubernetes_cluster.example.id
  vm_size               = "Basic_A2"
}

resource "azurerm_kubernetes_cluster_node_pool" "with_min_count" {
  name                  = "internal"
  kubernetes_cluster_id = azurerm_kubernetes_cluster.example.id
  vm_size               = "Standard_DS2_v2"
  min_count             = 2
  os_disk_type          = "Ephemeral"
}

resource "azurerm_kubernetes_cluster_node_pool" "zero_min_count_default_node_count" {
  name                  = "internal"
  kubernetes_cluster_id = azurerm_kubernetes_cluster.example.id
  vm_size               = "Standard_DS2_v2"
  node_count            = 1
  min_count             = 0
  max_count             = 3
}

resource "azurerm_kubernetes_cluster_node_pool" "windows" {
  name                  = "internal"
  kubernetes_cluster_id = azurerm_kubernetes_cluster.example.id
  vm_size               = "Basic_A2"
  os_type               = "Windows"
}

resource "azurerm_kubernetes_cluster_node_pool" "windows_sku" {
  name                  = "internal"
  kubernetes_cluster_id = azurerm_kubernetes_cluster.example.id
  vm_size               = "Standard_DS2_v2"
  os_sku                = "Windows2022"
}

resource "azurerm_kubernetes_cluster_node_pool" "non_premium" {
  name                  = "internal"
  kubernetes_cluster_id = azurerm_kubernetes_cluster.example.id
  vm_size               = "Standard_D2_v3"
}
