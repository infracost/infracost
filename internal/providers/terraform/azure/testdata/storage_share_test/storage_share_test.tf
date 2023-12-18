provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "westus"
}

locals {
  share_premium_options = [
    {
      access_tiers              = ["Premium"]
      account_replication_types = ["LRS", "ZRS"],
    },
  ]

  share_standard_options = [
    {
      access_tiers              = ["Cool", "Hot", "TransactionOptimized"]
      account_replication_types = ["LRS", "GRS", "ZRS"],
    },
  ]

  storage_account_premium_permutations = distinct(flatten([
    for share_premium_option in local.share_premium_options : [
      for account_replication_type in share_premium_option.account_replication_types : {
        account_replication_type = account_replication_type
      }
    ]
  ]))

  share_premium_permuations = distinct(flatten([
    for share_premium_option in local.share_premium_options : [
      for account_replication_type in share_premium_option.account_replication_types : [
        for access_tier in share_premium_option.access_tiers : {
          access_tier              = access_tier
          account_replication_type = account_replication_type
        }
      ]
    ]
  ]))

  storage_account_standard_permutations = distinct(flatten([
    for share_standard_option in local.share_standard_options : [
      for account_replication_type in share_standard_option.account_replication_types : {
        account_replication_type = account_replication_type
      }
    ]
  ]))

  share_standard_permuations = distinct(flatten([
    for share_standard_option in local.share_standard_options : [
      for account_replication_type in share_standard_option.account_replication_types : [
        for access_tier in share_standard_option.access_tiers : {
          access_tier              = access_tier
          account_replication_type = account_replication_type
        }
      ]
    ]
  ]))
}

resource "azurerm_storage_account" "filestorage" {
  for_each = { for entry in local.storage_account_premium_permutations : "${entry.account_replication_type}" => entry }

  name                     = substr(lower("icfilestorage${each.value.account_replication_type}"), 0, 24)
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "FileStorage"
  account_tier             = "Premium"
  account_replication_type = each.value.account_replication_type
}

resource "azurerm_storage_share" "premium-filestorage" {
  for_each = { for entry in local.share_premium_permuations : "${entry.access_tier}.${entry.account_replication_type}" => entry }

  name                 = substr(lower("premium${each.value.access_tier}${each.value.account_replication_type}"), 0, 24)
  storage_account_name = azurerm_storage_account.filestorage["${each.value.account_replication_type}"].name
  quota                = 50
  access_tier          = each.value.access_tier
}

resource "azurerm_storage_account" "standard" {
  for_each = { for entry in local.storage_account_standard_permutations : "${entry.account_replication_type}" => entry }

  name                     = substr(lower("icstandard${each.value.account_replication_type}"), 0, 24)
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "StorageV2"
  account_tier             = "Standard"
  account_replication_type = each.value.account_replication_type
}

resource "azurerm_storage_share" "standard" {
  for_each = { for entry in local.share_standard_permuations : "${entry.access_tier}.${entry.account_replication_type}" => entry }

  name                 = substr(lower("standard${each.value.access_tier}${each.value.account_replication_type}"), 0, 24)
  storage_account_name = azurerm_storage_account.standard["${each.value.account_replication_type}"].name
  quota                = 50
  access_tier          = each.value.access_tier
}

resource "azurerm_storage_account" "premium" {
  name                     = "icpremium"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "StorageV2"
  account_tier             = "Premium"
  account_replication_type = "LRS"
}

resource "azurerm_storage_share" "unsupported" {
  name                 = "unsupported"
  storage_account_name = azurerm_storage_account.premium.name
  quota                = 50
  access_tier          = "Premium"
}
