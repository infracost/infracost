provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_traffic_manager_profile" "example" {
  name                   = "example-profile"
  profile_status         = "Enabled"
  resource_group_name    = azurerm_resource_group.example.name
  traffic_routing_method = "Weighted"

  dns_config {
    relative_name = "example-profile"
    ttl           = 100
  }

  monitor_config {
    protocol = "HTTP"
    port     = 80
  }
}

resource "azurerm_traffic_manager_profile" "example_with_usage" {
  name                   = "example-profile"
  profile_status         = "Enabled"
  resource_group_name    = azurerm_resource_group.example.name
  traffic_routing_method = "Weighted"
  traffic_view_enabled   = true

  dns_config {
    relative_name = "example-profile"
    ttl           = 100
  }

  monitor_config {
    protocol = "HTTP"
    port     = 80
  }
}


resource "azurerm_traffic_manager_profile" "example_with_traffic_view" {
  name                   = "example-profile-with_traffic_view"
  resource_group_name    = azurerm_resource_group.example.name
  traffic_routing_method = "Weighted"
  traffic_view_enabled   = true

  dns_config {
    relative_name = "example-profile"
    ttl           = 100
  }

  monitor_config {
    protocol = "HTTP"
    port     = 80
  }
}

resource "azurerm_traffic_manager_profile" "example_disabled" {
  name                   = "example-profile"
  profile_status         = "Disabled"
  resource_group_name    = azurerm_resource_group.example.name
  traffic_routing_method = "Weighted"
  traffic_view_enabled   = true

  dns_config {
    relative_name = "example-profile"
    ttl           = 100
  }

  monitor_config {
    protocol = "HTTP"
    port     = 80
  }
}
