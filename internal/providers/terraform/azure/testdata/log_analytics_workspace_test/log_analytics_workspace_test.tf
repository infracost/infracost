provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_log_analytics_workspace" "free_workspace" {
  name                = "acctest-free"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "Free"
}

resource "azurerm_log_analytics_workspace" "per_gb_data_ingestion" {
  name                = "acctest-01"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "PerGB2018"
}

resource "azurerm_log_analytics_workspace" "capacity_gb_data_ingestion" {
  name                               = "acctest-02"
  location                           = azurerm_resource_group.example.location
  resource_group_name                = azurerm_resource_group.example.name
  sku                                = "CapacityReservation"
  reservation_capacity_in_gb_per_day = 100
}

resource "azurerm_log_analytics_workspace" "capacity_gb_data_above_commitment_tiers" {
  name                = "acctest-09"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "CapacityReservation"
  # this reservation capacity doesn't actually exist in the azure pricing page. It should appear in the golden file as
  # 1000 gb a day commitment tier.
  reservation_capacity_in_gb_per_day = 1200
}

resource "azurerm_log_analytics_workspace" "capacity_gb_data_ingestion_without_specification" {
  name                = "acctest-03"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "CapacityReservation"
}

resource "azurerm_log_analytics_workspace" "log_data_retention_free" {
  name                = "acctest-04"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "PerGB2018"
  retention_in_days   = 30
}

resource "azurerm_log_analytics_workspace" "log_data_retention_with_usage" {
  name                = "acctest-05"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "PerGB2018"
  retention_in_days   = 33
}

resource "azurerm_log_analytics_workspace" "log_data_retention_without_usage" {
  name                = "acctest-08"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "PerGB2018"
  retention_in_days   = 33
}

resource "azurerm_log_analytics_workspace" "log_data_export" {
  name                = "acctest-06"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "PerGB2018"
}


resource "azurerm_log_analytics_workspace" "unsupported_legacy_workspace" {
  for_each            = toset(["Unlimited", "Standard", "Premium", "PerNode"])
  name                = "acctest-unsupported-${each.key}"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = each.value
}

