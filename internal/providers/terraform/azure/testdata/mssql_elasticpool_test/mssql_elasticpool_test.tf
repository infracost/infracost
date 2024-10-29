provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_mssql_server" "example" {
  name                         = "myexamplesqlserver"
  resource_group_name          = azurerm_resource_group.example.name
  location                     = "eastus"
  version                      = "12.0"
  administrator_login          = "admin"
  administrator_login_password = "password"
}

resource "azurerm_mssql_elasticpool" "gp_gen5" {
  name                = "gp-gen5"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  server_name         = azurerm_mssql_server.example.name
  license_type        = "LicenseIncluded"
  max_size_gb         = 100

  sku {
    name     = "GP_Gen5"
    tier     = "GeneralPurpose"
    family   = "Gen5"
    capacity = 4
  }

  per_database_settings {
    min_capacity = 0.25
    max_capacity = 4
  }
}

resource "azurerm_mssql_elasticpool" "gp_gen5_zone_redundant" {
  name                = "gp-gen5"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  server_name         = azurerm_mssql_server.example.name
  license_type        = "LicenseIncluded"
  zone_redundant      = true
  max_size_gb         = 100

  sku {
    name     = "GP_Gen5"
    tier     = "GeneralPurpose"
    family   = "Gen5"
    capacity = 4
  }

  per_database_settings {
    min_capacity = 0.25
    max_capacity = 4
  }
}

resource "azurerm_mssql_elasticpool" "gp_gen5_zone_no_license" {
  name                = "gp-gen5"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  server_name         = azurerm_mssql_server.example.name
  license_type        = "BasePrice"
  max_size_gb         = 100

  sku {
    name     = "GP_Gen5"
    tier     = "GeneralPurpose"
    family   = "Gen5"
    capacity = 4
  }

  per_database_settings {
    min_capacity = 0.25
    max_capacity = 4
  }
}

resource "azurerm_mssql_elasticpool" "bc_dc" {
  name                = "bc-dc"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  server_name         = azurerm_mssql_server.example.name
  license_type        = "LicenseIncluded"
  max_size_gb         = 100

  sku {
    name     = "BC_DC"
    tier     = "BusinessCritical"
    family   = "DC"
    capacity = 8
  }

  per_database_settings {
    min_capacity = 0.25
    max_capacity = 4
  }
}

resource "azurerm_mssql_elasticpool" "bc_dc_zone_redundant" {
  name                = "bc-dc-zone-redundant"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  server_name         = azurerm_mssql_server.example.name
  license_type        = "LicenseIncluded"
  max_size_gb         = 100
  zone_redundant      = true

  sku {
    name     = "BC_DC"
    tier     = "BusinessCritical"
    family   = "DC"
    capacity = 8
  }

  per_database_settings {
    min_capacity = 0.25
    max_capacity = 4
  }
}

resource "azurerm_mssql_elasticpool" "basic_100" {
  name                = "basic-100"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  server_name         = azurerm_mssql_server.example.name
  max_size_gb         = 9.7656250

  sku {
    name     = "BasicPool"
    tier     = "Basic"
    capacity = 100
  }

  per_database_settings {
    min_capacity = 1
    max_capacity = 4
  }
}

resource "azurerm_mssql_elasticpool" "standard_200" {
  name                = "standard-200"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  server_name         = azurerm_mssql_server.example.name
  max_size_gb         = 300 # 100 GB extra storage

  sku {
    name     = "StandardPool"
    tier     = "Standard"
    capacity = 200
  }

  per_database_settings {
    min_capacity = 1
    max_capacity = 4
  }
}

resource "azurerm_mssql_elasticpool" "premium_500" {
  name                = "premium-500"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  server_name         = azurerm_mssql_server.example.name
  max_size_gb         = 1024 # 274 GB extra storage

  sku {
    name     = "PremiumPool"
    tier     = "Premium"
    capacity = 500
  }

  per_database_settings {
    min_capacity = 1
    max_capacity = 4
  }
}
