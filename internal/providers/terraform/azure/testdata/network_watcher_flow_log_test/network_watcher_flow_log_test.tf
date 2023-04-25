provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

locals {
  permutations = {
    "disabled" = {
      enabled = false,
    },
    "traffic-analytics-disabled" = {
      traffic_analytics_enabled = false,
    },
    "traffic-analytics-interval-10" = {
      traffic_analytics_interval_in_minutes = 10,
    },
    "traffic-analytics-interval-60" = {
      traffic_analytics_interval_in_minutes = 60,
    }
  }
}

resource "azurerm_network_watcher_flow_log" "network_watcher_flow_log" {
  for_each = { for key, value in local.permutations : key => value }

  name                 = "network-watcher-flow-log"
  network_watcher_name = "network-watcher-name"
  resource_group_name  = azurerm_resource_group.example.name

  network_security_group_id = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Network/networkSecurityGroups/testecuritygroup"
  storage_account_id        = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Storage/storageAccounts/teststorageaccount"
  enabled                   = lookup(each.value, "enabled", true)

  retention_policy {
    enabled = true
    days    = 7
  }

  traffic_analytics {
    enabled               = lookup(each.value, "traffic_analytics_enabled", true)
    workspace_id          = "00000000-0000-0000-0000-000000000000"
    workspace_region      = "westeurope"
    workspace_resource_id = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.OperationalInsights/workspaces/testworkspace"
    interval_in_minutes   = lookup(each.value, "traffic_analytics_interval_in_minutes", 60)
  }
}

resource "azurerm_network_watcher_flow_log" "network_watcher_flow_log_with_usage" {
  for_each = { for key, value in local.permutations : key => value }

  name                 = "network-watcher-flow-log"
  network_watcher_name = "network-watcher-name"
  resource_group_name  = azurerm_resource_group.example.name

  network_security_group_id = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Network/networkSecurityGroups/testecuritygroup"
  storage_account_id        = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Storage/storageAccounts/teststorageaccount"
  enabled                   = lookup(each.value, "enabled", true)

  retention_policy {
    enabled = true
    days    = 7
  }

  traffic_analytics {
    enabled               = lookup(each.value, "traffic_analytics_enabled", true)
    workspace_id          = "00000000-0000-0000-0000-000000000000"
    workspace_region      = "westeurope"
    workspace_resource_id = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.OperationalInsights/workspaces/testworkspace"
    interval_in_minutes   = lookup(each.value, "traffic_analytics_interval_in_minutes", 60)
  }
}

