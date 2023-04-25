provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_security_center_subscription_pricing" "free_example" {
  tier = "Free"
}

resource "azurerm_security_center_subscription_pricing" "default_resource_type_example" {
  tier = "Standard"
}

resource "azurerm_security_center_subscription_pricing" "standard_example" {
  for_each = toset(["AppServices", "ContainerRegistry", "KeyVaults", "KubernetesService", "SqlServers", "SqlServerVirtualMachines", "StorageAccounts", "VirtualMachines", "Arm", "Dns", "OpenSourceRelationalDatabases", "Containers", "CosmosDbs", "CloudPosture"])

  tier          = "Standard"
  resource_type = each.key
}

resource "azurerm_security_center_subscription_pricing" "standard_example_with_usage" {
  for_each = toset(["AppServices", "ContainerRegistry", "KeyVaults", "KubernetesService", "SqlServers", "SqlServerVirtualMachines", "StorageAccounts", "VirtualMachines", "Arm", "Dns", "OpenSourceRelationalDatabases", "Containers", "CosmosDbs", "CloudPosture"])

  tier          = "Standard"
  resource_type = each.key
}
