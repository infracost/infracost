provider "azurerm" {
  features {}
  skip_provider_registration = true
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_key_vault" "premium" {
  name                       = "examplekeyvault"
  location                   = "eastus"
  resource_group_name        = azurerm_resource_group.example.name
  tenant_id                  = "00000000-0000-0000-0000-000000000000"
  sku_name                   = "premium"
  soft_delete_retention_days = 7
}

resource "azurerm_key_vault" "standatd" {
  name                       = "examplekeyvault"
  location                   = "eastus"
  resource_group_name        = azurerm_resource_group.example.name
  tenant_id                  = "00000000-0000-0000-0000-000000000000"
  sku_name                   = "standard"
  soft_delete_retention_days = 7
}

resource "azurerm_key_vault_key" "pr_rsa2048" {
  name         = "generated-certificate"
  key_vault_id = azurerm_key_vault.premium.id
  key_type     = "RSA"
  key_size     = 2048

  key_opts = [
  ]
}

resource "azurerm_key_vault_key" "pr_rsa3072" {
  name         = "generated-certificate"
  key_vault_id = azurerm_key_vault.premium.id
  key_type     = "RSA"
  key_size     = 3072

  key_opts = [
  ]
}

resource "azurerm_key_vault_key" "pr_ec" {
  name         = "generated-certificate"
  key_vault_id = azurerm_key_vault.premium.id
  key_type     = "EC"

  key_opts = [
  ]
}

resource "azurerm_key_vault_key" "pr_hsm_rsa3072" {
  name         = "generated-certificate"
  key_vault_id = azurerm_key_vault.premium.id
  key_type     = "RSA-HSM"
  key_size     = 3072

  key_opts = [
  ]
}

resource "azurerm_key_vault_key" "pr_hsm_rsa2048" {
  name         = "generated-certificate"
  key_vault_id = azurerm_key_vault.premium.id
  key_type     = "RSA-HSM"
  key_size     = 2048

  key_opts = [
  ]
}

resource "azurerm_key_vault_key" "std_hsm_rsa3072" {
  name         = "generated-certificate"
  key_vault_id = azurerm_key_vault.standatd.id
  key_type     = "RSA"
  key_size     = 3072

  key_opts = [
  ]
}

resource "azurerm_key_vault_key" "std_rsa3072" {
  name         = "generated-certificate"
  key_vault_id = azurerm_key_vault.standatd.id
  key_type     = "RSA"
  key_size     = 3072

  key_opts = [
  ]
}

resource "azurerm_key_vault_key" "std_rsa2048" {
  name         = "generated-certificate"
  key_vault_id = azurerm_key_vault.standatd.id
  key_type     = "RSA"
  key_size     = 3072

  key_opts = [
  ]
}

resource "azurerm_key_vault_key" "no_usage" {
  name         = "generated-certificate"
  key_vault_id = azurerm_key_vault.premium.id
  key_type     = "EC"

  key_opts = [
  ]
}
