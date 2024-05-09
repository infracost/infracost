provider "azurerm" {
  skip_provider_registration = true
  features {}
}

locals {
  skus = [
    "free", "standard", ""
  ]
  replicas = [
    [],
    [
      {
        name     = "us"
        location = "West US"
      }
    ],
    [
      {
        name     = "us"
        location = "West US"
      },
      {
        name     = "asia"
        location = "East Asia"
      },
      {
        name     = "au"
        location = "Australia East"
      }
    ]
  ]
  permutations = distinct(flatten([
    for replicas in local.replicas : [
      for sku in local.skus : {
        replicas = replicas
        sku      = sku
      }
    ]
  ]))
}
resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_app_configuration" "base" {
  for_each = { for entry in local.permutations : "${entry.sku}.replicas.${length(entry.replicas)}" => entry }

  sku                 = each.value.sku
  name                = "${each.value.sku}.replicas.${length(each.value.replicas)}"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location

  dynamic "replica" {
    for_each = each.value.replicas
    content {
      name     = replica.value.name
      location = replica.value.location
    }
  }
}

resource "azurerm_app_configuration" "usage" {
  for_each = { for entry in local.permutations : "${entry.sku}.replicas.${length(entry.replicas)}" => entry }

  sku                 = each.value.sku
  name                = "${each.value.sku}-r-${length(each.value.replicas)}"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location

  dynamic "replica" {
    for_each = each.value.replicas
    content {
      name     = replica.value.name
      location = replica.value.location
    }
  }
}
