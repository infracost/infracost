provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

locals {
  options = [
    {
      sku      = ["Basic", "Standard"]
      capacity = [0]
    },
    {
      sku      = ["Premium"]
      capacity = [1, 2, 4, 8, 16]
    },
  ]

  permutations = distinct(flatten([
    for option in local.options : [
      for sku in option.sku : [
        for capacity in option.capacity : {
          sku      = sku
          capacity = capacity
        }
      ]
    ]
  ]))
}

resource "azurerm_servicebus_namespace" "servicebus_namespace" {
  for_each = { for entry in local.permutations : "${entry.sku}.${entry.capacity}" => entry }

  name                = "servicebus-namespace"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location

  sku      = each.value.sku
  capacity = each.value.capacity
}

resource "azurerm_servicebus_namespace" "servicebus_namespace_with_usage" {
  for_each = { for entry in local.permutations : "${entry.sku}.${entry.capacity}" => entry }

  name                = "servicebus-namespace-with-usage"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location

  sku      = each.value.sku
  capacity = each.value.capacity
}
