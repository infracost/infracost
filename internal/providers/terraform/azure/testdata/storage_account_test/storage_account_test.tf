provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "westus"
}

locals {
  blockblob_options = [
    {
      account_kind              = "BlockBlobStorage"
      account_tier              = "Standard"
      account_replication_types = ["LRS", "GRS", "RAGRS"]
    },
    {
      account_kind              = "BlockBlobStorage"
      account_tier              = "Premium"
      account_replication_types = ["LRS", "ZRS"]
    }
  ]

  blob_options = [
    {
      account_kind              = "BlobStorage"
      account_tier              = "Standard"
      account_replication_types = ["LRS", "GRS"]
      access_tiers              = ["Hot", "Cool"]
    },
    {
      account_kind              = "BlobStorage"
      account_tier              = "Premium"
      account_replication_types = ["LRS", "ZRS"]
      access_tiers              = ["Hot", "Cool"]
    }
  ]

  storagev1_options = [
    {
      account_kind              = "Storage"
      account_tier              = "Standard"
      account_replication_types = ["LRS", "ZRS", "GRS", "RAGRS"]
    }
  ]

  storagev2_options = [
    {
      account_kind              = "StorageV2"
      account_tier              = "Standard"
      account_replication_types = ["LRS", "GRS", "RAGRS", "ZRS", "GZRS", "RAGZRS"],
      access_tiers              = ["Hot", "Cool"]
    },
    {
      account_kind              = "StorageV2"
      account_tier              = "Premium"
      account_replication_types = ["LRS", "ZRS"]
      access_tiers              = ["Hot", "Cool"]
    }
  ]

  file_options = [
    {
      account_kind              = "FileStorage"
      account_tier              = "Standard"
      account_replication_types = ["LRS", "GRS", "ZRS"]
      access_tiers              = ["Hot", "Cool"]
    },
    {
      account_kind              = "FileStorage"
      account_tier              = "Premium"
      account_replication_types = ["LRS", "ZRS"]
      access_tiers              = ["Hot", "Cool"]
    }
  ]

  nfsv3_storagev2_options = [
    {
      account_kind              = "StorageV2"
      account_tier              = "Standard"
      account_replication_types = ["LRS", "GRS"]
      access_tiers              = ["Hot", "Cool"]
    }
  ]

  nfsv3_blockblob_options = [
    {
      account_kind              = "BlockBlobStorage"
      account_tier              = "Premium"
      account_replication_types = ["LRS", "ZRS"]
    }
  ]

  blockblob_permutations = distinct(flatten([
    for blockblob_option in local.blockblob_options : [
      for account_replication_type in blockblob_option.account_replication_types : {
        account_kind             = blockblob_option.account_kind
        account_tier             = blockblob_option.account_tier
        account_replication_type = account_replication_type
      }
    ]
  ]))

  blob_permutations = distinct(flatten([
    for blob_option in local.blob_options : [
      for account_replication_type in blob_option.account_replication_types : [
        for access_tier in blob_option.access_tiers : {
          account_kind             = blob_option.account_kind
          account_tier             = blob_option.account_tier
          account_replication_type = account_replication_type
          access_tier              = access_tier
        }
      ]
    ]
  ]))

  storagev1_permutations = distinct(flatten([
    for storagev1_option in local.storagev1_options : [
      for account_replication_type in storagev1_option.account_replication_types : {
        account_kind             = storagev1_option.account_kind
        account_tier             = storagev1_option.account_tier
        account_replication_type = account_replication_type
      }
    ]
  ]))

  storagev2_permutations = distinct(flatten([
    for storagev2_option in local.storagev2_options : [
      for account_replication_type in storagev2_option.account_replication_types : [
        for access_tier in storagev2_option.access_tiers : {
          account_kind             = storagev2_option.account_kind
          account_tier             = storagev2_option.account_tier
          account_replication_type = account_replication_type
          access_tier              = access_tier
        }
      ]
    ]
  ]))

  file_permutations = distinct(flatten([
    for file_option in local.file_options : [
      for account_replication_type in file_option.account_replication_types : [
        for access_tier in file_option.access_tiers : {
          account_kind             = file_option.account_kind
          account_tier             = file_option.account_tier
          account_replication_type = account_replication_type
          access_tier              = access_tier
        }
      ]
    ]
  ]))

  nfsv3_storagev2_permutations = distinct(flatten([
    for nfsv3_option in local.nfsv3_storagev2_options : [
      for account_replication_type in nfsv3_option.account_replication_types : [
        for access_tier in nfsv3_option.access_tiers : {
          account_kind             = nfsv3_option.account_kind
          account_tier             = nfsv3_option.account_tier
          account_replication_type = account_replication_type
          access_tier              = access_tier
        }
      ]
    ]
  ]))

  nfsv3_blockblob_permutations = distinct(flatten([
    for nfsv3_option in local.nfsv3_blockblob_options : [
      for account_replication_type in nfsv3_option.account_replication_types : {
        account_kind             = nfsv3_option.account_kind
        account_tier             = nfsv3_option.account_tier
        account_replication_type = account_replication_type
      }
    ]
  ]))
}

resource "azurerm_storage_account" "blockblob" {
  for_each = { for entry in local.blockblob_permutations : "${entry.account_kind}.${entry.account_tier}.${entry.account_replication_type}" => entry }

  name                     = substr(lower("ic${each.value.account_kind}${each.value.account_tier}${each.value.account_replication_type}"), 0, 24)
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = each.value.account_kind
  account_tier             = each.value.account_tier
  account_replication_type = each.value.account_replication_type
}

resource "azurerm_storage_account" "blob" {
  for_each = { for entry in local.blob_permutations : "${entry.account_kind}.${entry.account_tier}.${entry.account_replication_type}.${entry.access_tier}" => entry }

  name                     = substr(lower("ic${each.value.account_kind}${each.value.account_tier}${each.value.account_replication_type}${each.value.access_tier}"), 0, 24)
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = each.value.account_kind
  account_tier             = each.value.account_tier
  account_replication_type = each.value.account_replication_type
  access_tier              = each.value.access_tier
}

resource "azurerm_storage_account" "storagev1" {
  for_each = { for entry in local.storagev1_permutations : "${entry.account_kind}.${entry.account_tier}.${entry.account_replication_type}" => entry }

  name                     = substr(lower("ic${each.value.account_kind}${each.value.account_tier}${each.value.account_replication_type}"), 0, 24)
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = each.value.account_kind
  account_tier             = each.value.account_tier
  account_replication_type = each.value.account_replication_type
}

resource "azurerm_storage_account" "storagev2" {
  for_each = { for entry in local.storagev2_permutations : "${entry.account_kind}.${entry.account_tier}.${entry.account_replication_type}.${entry.access_tier}" => entry }

  name                     = substr(lower("ic${each.value.account_kind}${each.value.account_tier}${each.value.account_replication_type}${each.value.access_tier}"), 0, 24)
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = each.value.account_kind
  account_tier             = each.value.account_tier
  account_replication_type = each.value.account_replication_type
  access_tier              = each.value.access_tier
}

resource "azurerm_storage_account" "file" {
  for_each = { for entry in local.file_permutations : "${entry.account_kind}.${entry.account_tier}.${entry.account_replication_type}.${entry.access_tier}" => entry }

  name                     = substr(lower("ic${each.value.account_kind}${each.value.account_tier}${each.value.account_replication_type}${each.value.access_tier}"), 0, 24)
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = each.value.account_kind
  account_tier             = each.value.account_tier
  account_replication_type = each.value.account_replication_type
  access_tier              = each.value.access_tier
}

resource "azurerm_storage_account" "nfsv3_storagev2" {
  for_each = { for entry in local.nfsv3_storagev2_permutations : "${entry.account_kind}.${entry.account_tier}.${entry.account_replication_type}.${entry.access_tier}" => entry }

  name                     = substr(lower("ic${each.value.account_kind}${each.value.account_tier}${each.value.account_replication_type}${each.value.access_tier}"), 0, 24)
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = each.value.account_kind
  account_tier             = each.value.account_tier
  account_replication_type = each.value.account_replication_type
  access_tier              = each.value.access_tier
  nfsv3_enabled            = true
}

resource "azurerm_storage_account" "nfsv3_blockblob" {
  for_each = { for entry in local.nfsv3_blockblob_permutations : "${entry.account_kind}.${entry.account_tier}.${entry.account_replication_type}" => entry }

  name                     = substr(lower("ic${each.value.account_kind}${each.value.account_tier}${each.value.account_replication_type}"), 0, 24)
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = each.value.account_kind
  account_tier             = each.value.account_tier
  account_replication_type = each.value.account_replication_type
  nfsv3_enabled            = true
}

resource "azurerm_storage_account" "unsupported" {
  name                     = "icunsupported"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "BlockBlobStorage"
  account_tier             = "Premium"
  account_replication_type = "GRS"
}
