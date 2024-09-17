provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "westus"
}

resource "azurerm_storage_account" "unsupported" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "BlobStorage"
  account_tier             = "Premium"
  account_replication_type = "ZRS"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "bb_Premium_unsupported_access_tier" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "BlockBlobStorage"
  account_tier             = "Premium"
  account_replication_type = "RAGZRS"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "bb_Premium_ZRS" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "BlockBlobStorage"
  account_tier             = "Premium"
  account_replication_type = "ZRS"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "bb_Premium_LRS" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "BlockBlobStorage"
  account_tier             = "Premium"
  account_replication_type = "LRS"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "bb_Standard_unsupported_access_tier" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "BlockBlobStorage"
  account_tier             = "Standard"
  account_replication_type = "RAGZRS"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "bb_Standard_LRS_Hot" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "BlockBlobStorage"
  account_tier             = "Standard"
  account_replication_type = "LRS"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "bb_Standard_LRS_Cool" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "BlockBlobStorage"
  account_tier             = "Standard"
  account_replication_type = "LRS"
  access_tier              = "Cool"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "bb_Standard_GRS_Hot" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "BlockBlobStorage"
  account_tier             = "Standard"
  account_replication_type = "GRS"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "bb_Standard_GRS_Cool" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "BlockBlobStorage"
  account_tier             = "Standard"
  account_replication_type = "GRS"
  access_tier              = "Cool"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "bb_Standard_RAGRS_Hot" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "BlockBlobStorage"
  account_tier             = "Standard"
  account_replication_type = "RAGRS"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "bb_Standard_RAGRS_Cool" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "BlockBlobStorage"
  account_tier             = "Standard"
  account_replication_type = "RAGRS"
  access_tier              = "Cool"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "bb_without_usage_file" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "BlockBlobStorage"
  account_tier             = "Standard"
  account_replication_type = "RAGRS"
  access_tier              = "Cool"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "file_without_usage_file" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "FileStorage"
  account_tier             = "Standard"
  account_replication_type = "LRS"
  access_tier              = "Cool"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "file_unsupported_access_tier" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "FileStorage"
  account_tier             = "Standard"
  account_replication_type = "GZRS"
  access_tier              = "Cool"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "file_cool_lrs" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "FileStorage"
  account_tier             = "Standard"
  account_replication_type = "LRS"
  access_tier              = "Cool"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "file_hot_lrs" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "FileStorage"
  account_tier             = "Standard"
  account_replication_type = "LRS"
  access_tier              = "Hot"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "file_cool_grs" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "FileStorage"
  account_tier             = "Standard"
  account_replication_type = "GRS"
  access_tier              = "Cool"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "file_hot_grs" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "FileStorage"
  account_tier             = "Standard"
  account_replication_type = "GRS"
  access_tier              = "Hot"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "file_premium_unsupported_access_tier" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "FileStorage"
  account_tier             = "Premium"
  account_replication_type = "GZRS"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "file_premium_lrs" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "FileStorage"
  account_tier             = "Premium"
  account_replication_type = "LRS"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "file_premium_zrs" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "FileStorage"
  account_tier             = "Premium"
  account_replication_type = "ZRS"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "v2_without_usage" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "StorageV2"
  account_tier             = "Standard"
  account_replication_type = "LRS"
  access_tier              = "Hot"
  nfsv3_enabled            = true
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "v2_cool_lrs" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "StorageV2"
  account_tier             = "Standard"
  account_replication_type = "LRS"
  access_tier              = "Cool"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "v2_cool_lrs_nfsv3" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "StorageV2"
  account_tier             = "Standard"
  account_replication_type = "LRS"
  access_tier              = "Cool"
  nfsv3_enabled            = true
  min_tls_version          = "TLS1_2"
}

# Temporarily disabled due to duplication issue in Azure StorageV2 pricing
# records (same productHash for different cost components).
# resource "azurerm_storage_account" "v2_hot_lrs" {
#   name                     = "storageaccountname"
#   resource_group_name      = azurerm_resource_group.example.name
#   location                 = azurerm_resource_group.example.location
#   account_kind             = "StorageV2"
#   account_tier             = "Standard"
#   account_replication_type = "LRS"
#   access_tier              = "Hot"
# }

resource "azurerm_storage_account" "v2_hot_lrs_nfsv3" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "StorageV2"
  account_tier             = "Standard"
  account_replication_type = "LRS"
  access_tier              = "Hot"
  nfsv3_enabled            = true
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "v2_cool_gzrs" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "StorageV2"
  account_tier             = "Standard"
  account_replication_type = "GZRS"
  access_tier              = "Cool"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "v2_hot_gzrs" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "StorageV2"
  account_tier             = "Standard"
  account_replication_type = "GZRS"
  access_tier              = "Hot"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "v2_cool_ragzrs" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "StorageV2"
  account_tier             = "Standard"
  account_replication_type = "RAGZRS"
  access_tier              = "Cool"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "v2_hot_ragzrs" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "StorageV2"
  account_tier             = "Standard"
  account_replication_type = "RAGZRS"
  access_tier              = "Hot"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "v2_premium_unsupported_access_tier" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "StorageV2"
  account_tier             = "Premium"
  account_replication_type = "GZRS"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "v2_premium_lrs" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "StorageV2"
  account_tier             = "Premium"
  account_replication_type = "LRS"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account" "v2_premium_lrs_nfsv3" {
  name                     = "storageaccountname"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_kind             = "StorageV2"
  account_tier             = "Premium"
  account_replication_type = "LRS"
  nfsv3_enabled            = true
  min_tls_version          = "TLS1_2"
}
