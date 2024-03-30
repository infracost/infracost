provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

locals {
  kind_skus = {
    "SpeechServices" : ["F0", "S0"],
    "LUIS" : ["F0", "S0"],
    "TextAnalytics" : ["F0", "S0"],
  }

  permutations = distinct(flatten([
    for kind, skus in local.kind_skus : [
      for sku in skus : {
        kind = kind
        sku  = sku
      }
    ]
  ]))
}

resource "azurerm_cognitive_account" "without_usage" {
  for_each = { for perm in local.permutations : "${perm.kind}-${perm.sku}" => perm }

  name                = "without-usage-${each.key}"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  kind                = each.value.kind
  sku_name            = each.value.sku
}

resource "azurerm_cognitive_account" "with_usage" {
  for_each = { for perm in local.permutations : "${perm.kind}-${perm.sku}" => perm }

  name                = "with-usage-${each.key}"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  kind                = each.value.kind
  sku_name            = each.value.sku
}

resource "azurerm_cognitive_account" "speech_with_commitment" {
  for_each = toset(["small", "medium", "large", "invalid"])

  name                = "speech-with-commitment-${each.key}"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  kind                = "SpeechServices"
  sku_name            = "S0"
}

resource "azurerm_cognitive_account" "luis_with_commitment" {
  for_each = toset(["small", "medium", "large", "invalid"])

  name                = "luis-with-commitment-${each.key}"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  kind                = "LUIS"
  sku_name            = "S0"
}

resource "azurerm_cognitive_account" "textanalytics_with_commitment" {
  for_each = toset(["small", "medium", "large", "invalid"])

  name                = "textanalytics-with-commitment-${each.key}"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  kind                = "TextAnalytics"
  sku_name            = "S0"
}

resource "azurerm_cognitive_account" "textanalytics_with_tiers" {
  name                = "textanalytics-with-tiers"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  kind                = "TextAnalytics"
  sku_name            = "S0"
}

resource "azurerm_cognitive_account" "unsupported" {
  name                = "unsupported"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  kind                = "Academic"
  sku_name            = "S0"
}
