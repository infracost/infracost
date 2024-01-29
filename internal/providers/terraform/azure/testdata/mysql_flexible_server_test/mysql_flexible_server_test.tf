provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "westus"
}

resource "azurerm_mysql_flexible_server" "gp" {
  name                = "example-mysqlflexibleserver"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location

  sku_name = "GP_Standard_D4ds_v4"
}

resource "azurerm_mysql_flexible_server" "gp_dXads" {
  name                = "example-mysqlflexibleserver-d2ads"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location

  sku_name = "GP_Standard_D2ads_v5"
}

resource "azurerm_mysql_flexible_server" "mo" {
  name                = "example-mysqlflexibleserver"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location

  storage {
    iops    = 500
    size_gb = 20
  }

  sku_name = "MO_Standard_E4ds_v4"
}

resource "azurerm_mysql_flexible_server" "burstable" {
  name                = "example-mysqlflexibleserver"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location

  storage {
    iops    = 360
    size_gb = 30
  }

  sku_name = "B_Standard_B1ms"
}

resource "azurerm_mysql_flexible_server" "non_usage_gp" {
  name                = "example-mysqlflexibleserver"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location

  sku_name = "GP_Standard_D16ds_v4"
}
