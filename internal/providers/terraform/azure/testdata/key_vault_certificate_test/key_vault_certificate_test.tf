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

resource "azurerm_key_vault" "standard" {
  name                       = "examplekeyvault"
  location                   = "eastus"
  resource_group_name        = azurerm_resource_group.example.name
  tenant_id                  = "00000000-0000-0000-0000-000000000000"
  sku_name                   = "standard"
  soft_delete_retention_days = 7
}

resource "azurerm_key_vault_certificate" "non-usage" {
  name         = "imported-cert"
  key_vault_id = azurerm_key_vault.standard.id

  certificate_policy {
    issuer_parameters {
      name = "Self"
    }

    key_properties {
      exportable = true
      key_type   = "RSA"
      reuse_key  = false
    }

    secret_properties {
      content_type = "application/x-pkcs12"
    }
  }
}

resource "azurerm_key_vault_certificate" "standard" {
  name         = "imported-cert"
  key_vault_id = azurerm_key_vault.standard.id

  certificate_policy {
    issuer_parameters {
      name = "Self"
    }

    key_properties {
      exportable = true
      key_type   = "RSA"
      reuse_key  = false
    }

    secret_properties {
      content_type = "application/x-pkcs12"
    }
  }
}

resource "azurerm_key_vault_certificate" "premium" {
  name         = "imported-cert"
  key_vault_id = azurerm_key_vault.premium.id

  certificate_policy {
    issuer_parameters {
      name = "Self"
    }

    key_properties {
      exportable = true
      key_type   = "RSA"
      reuse_key  = false
    }

    secret_properties {
      content_type = "application/x-pkcs12"
    }
  }
}
