provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "exampleRG1"
  location = "eastus"
}

locals {
  os_types = ["Windows", "Linux", "WindowsContainer"]
  skus = [
    "B1",
    "B2",
    "B3",
    "D1",
    "F1",
    "I1",
    "I2",
    "I3",
    "I1v2",
    "I2v2",
    "I3v2",
    "I4v2",
    "I5v2",
    "I6v2",
    "P1v2",
    "P2v2",
    "P0v3",
    "P3v2",
    "P1v3",
    "P2v3",
    "P3v3",
    "P1mv3",
    "P2mv3",
    "P3mv3",
    "P4mv3",
    "P5mv3",
    "S1",
    "S2",
    "S3",
    "SHARED",
    "EP1",
    "EP2",
    "EP3",
    "WS1",
    "WS2",
    "WS3",
    "Y1"
  ]
  worker_counts = [1, 2, 3]

  permutations = distinct(flatten([
    for os_type in local.os_types : [
      for sku in local.skus : [
        for worker in local.worker_counts : {
          sku     = sku
          worker  = worker
          os_type = os_type
        }
      ]
    ]
  ]))
}

resource "azurerm_service_plan" "example" {
  for_each = { for entry in local.permutations : "${entry.os_type}.${entry.sku}.${entry.worker}" => entry }

  name                = "example-app-service-plan-${each.value.os_type}-${each.value.sku}-${each.value.worker}"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku_name            = each.value.sku
  os_type             = each.value.os_type
  worker_count        = each.value.worker
}
