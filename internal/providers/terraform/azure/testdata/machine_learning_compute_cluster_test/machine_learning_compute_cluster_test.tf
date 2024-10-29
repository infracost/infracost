provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

locals {
  vm_sizes = [
    "Standard_D2s_v3",
    "Standard_D4s_v3",
    "Standard_D8s_v3",
    "Standard_D16s_v3",
    "Standard_DS12_v2",
    "Standard_F4s_v2",
    "Standard_F8s_v2",
    "Standard_F16s_v2",
    "Standard_E4s_v3",
    "Standard_E8s_v3",
    "Standard_E16s_v3",
    "Standard_NC6",
    "Standard_NC12",
    "Standard_NC24",
    "Standard_NV6",
    "Standard_NV12",
    "Standard_NV24",
  ]

  min_node_counts = [0, 1, 4, 8, 16]

  permutations = distinct(flatten([
    for vm_size in local.vm_sizes : [
      for min_node_count in local.min_node_counts : {
        vm_size        = vm_size
        min_node_count = min_node_count
      }
    ]
  ]))
}

resource "azurerm_storage_account" "example" {
  name                     = "examplestorageaccount"
  location                 = azurerm_resource_group.example.location
  resource_group_name      = azurerm_resource_group.example.name
  account_tier             = "Standard"
  account_replication_type = "GRS"
}

resource "azurerm_machine_learning_workspace" "example" {
  name                    = "example-workspace"
  location                = azurerm_resource_group.example.location
  resource_group_name     = azurerm_resource_group.example.name
  key_vault_id            = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group/providers/Microsoft.KeyVault/vaults/example"
  application_insights_id = "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/example-resource-group/providers/Microsoft.Insights/components/example"

  identity {
    type = "SystemAssigned"
  }

  storage_account_id = azurerm_storage_account.example.id
}

resource "azurerm_machine_learning_compute_cluster" "example" {
  for_each = { for entry in local.permutations : "${entry.vm_size}.${entry.min_node_count}" => entry }

  name                          = "example-cluster"
  location                      = azurerm_resource_group.example.location
  machine_learning_workspace_id = azurerm_machine_learning_workspace.example.id
  vm_size                       = each.value.vm_size

  scale_settings {
    min_node_count                       = each.value.min_node_count
    max_node_count                       = each.value.min_node_count + 1
    scale_down_nodes_after_idle_duration = 30
  }

  vm_priority = "LowPriority"
}

resource "azurerm_machine_learning_compute_cluster" "with_monthly_hrs" {
  name                          = "with-monthly-hrs"
  location                      = azurerm_resource_group.example.location
  machine_learning_workspace_id = azurerm_machine_learning_workspace.example.id
  vm_size                       = "Standard_D2s_v3"

  scale_settings {
    min_node_count                       = 2
    max_node_count                       = 20
    scale_down_nodes_after_idle_duration = 30
  }

  vm_priority = "LowPriority"
}

resource "azurerm_machine_learning_compute_cluster" "with_monthly_hrs_zero_min_node_count" {
  name                          = "with-monthly-hrs"
  location                      = azurerm_resource_group.example.location
  machine_learning_workspace_id = azurerm_machine_learning_workspace.example.id
  vm_size                       = "Standard_D2s_v3"

  scale_settings {
    min_node_count                       = 0
    max_node_count                       = 20
    scale_down_nodes_after_idle_duration = 30
  }

  vm_priority = "LowPriority"
}

resource "azurerm_machine_learning_compute_cluster" "with_instances" {
  name                          = "with-monthly-hrs"
  location                      = azurerm_resource_group.example.location
  machine_learning_workspace_id = azurerm_machine_learning_workspace.example.id
  vm_size                       = "Standard_D2s_v3"

  scale_settings {
    min_node_count                       = 2
    max_node_count                       = 20
    scale_down_nodes_after_idle_duration = 30
  }

  vm_priority = "LowPriority"
}

resource "azurerm_machine_learning_compute_cluster" "with_instances_and_monthly_hrs" {
  name                          = "with-monthly-hrs"
  location                      = azurerm_resource_group.example.location
  machine_learning_workspace_id = azurerm_machine_learning_workspace.example.id
  vm_size                       = "Standard_D2s_v3"

  scale_settings {
    min_node_count                       = 2
    max_node_count                       = 20
    scale_down_nodes_after_idle_duration = 30
  }

  vm_priority = "LowPriority"
}

