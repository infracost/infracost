provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_managed_disk" "standard" {
  name                = "standard"
  resource_group_name = "fake_resource_group"
  location            = "eastus"

  create_option        = "Empty"
  storage_account_type = "Standard_LRS"
}

resource "azurerm_managed_disk" "premium" {
  name                = "premium"
  resource_group_name = "fake_resource_group"
  location            = "eastus"

  create_option        = "Empty"
  storage_account_type = "Premium_LRS"
}

resource "azurerm_managed_disk" "custom_size_ssd" {
  name                = "custom_size_ssd"
  resource_group_name = "fake_resource_group"
  location            = "eastus"

  create_option        = "Empty"
  storage_account_type = "StandardSSD_LRS"
  disk_size_gb         = 1000
}

resource "azurerm_managed_disk" "ultra" {
  name                = "ultra"
  resource_group_name = "fake_resource_group"
  location            = "eastus"

  create_option        = "Empty"
  storage_account_type = "UltraSSD_LRS"
  disk_size_gb         = 2000
  disk_iops_read_write = 4000
  disk_mbps_read_write = 20
}

resource "azurerm_managed_disk" "premiumv2_default" {
  name                = "premiumv2_default"
  resource_group_name = "fake_resource_group"
  location            = "eastus"

  create_option        = "Empty"
  storage_account_type = "PremiumV2_LRS"
  disk_size_gb         = 1024
}

resource "azurerm_managed_disk" "premiumv2_custom" {
  name                = "premiumv2_custom"
  resource_group_name = "fake_resource_group"
  location            = "eastus"

  create_option        = "Empty"
  storage_account_type = "PremiumV2_LRS"
  disk_size_gb         = 10752
  disk_iops_read_write = 6000
  disk_mbps_read_write = 500
  zone                 = "1"
}
