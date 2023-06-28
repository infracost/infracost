provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_traffic_manager_profile" "default_healthcheck_example" {
  name                   = "example-profile"
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

resource "azurerm_traffic_manager_external_endpoint" "default_healthcheck_example" {
  name       = "example-endpoint"
  profile_id = azurerm_traffic_manager_profile.default_healthcheck_example.id
  weight     = 100
  target     = "sometarget"
}


resource "azurerm_traffic_manager_profile" "basic_healthcheck_example" {
  name                   = "example-profile"
  resource_group_name    = azurerm_resource_group.example.name
  traffic_routing_method = "Weighted"

  dns_config {
    relative_name = "example-profile"
    ttl           = 100
  }

  monitor_config {
    protocol            = "HTTP"
    port                = 80
    interval_in_seconds = 30
  }
}

resource "azurerm_traffic_manager_external_endpoint" "basic_healthcheck_example" {
  name       = "example-endpoint"
  profile_id = azurerm_traffic_manager_profile.basic_healthcheck_example.id
  weight     = 100
  target     = "sometarget"
}

resource "azurerm_traffic_manager_profile" "fast_healthcheck_example" {
  name                   = "example-profile"
  resource_group_name    = azurerm_resource_group.example.name
  traffic_routing_method = "Weighted"

  dns_config {
    relative_name = "example-profile"
    ttl           = 100
  }

  monitor_config {
    protocol            = "HTTP"
    port                = 80
    interval_in_seconds = 10
  }
}

resource "azurerm_traffic_manager_external_endpoint" "fast_healthcheck_example" {
  name       = "example-endpoint"
  profile_id = azurerm_traffic_manager_profile.fast_healthcheck_example.id
  weight     = 100
  target     = "sometarget"
}

resource "azurerm_traffic_manager_profile" "disabled_example" {
  name                   = "example-profile"
  resource_group_name    = azurerm_resource_group.example.name
  traffic_routing_method = "Weighted"
  profile_status         = "Disabled"

  dns_config {
    relative_name = "example-profile"
    ttl           = 100
  }

  monitor_config {
    protocol            = "HTTP"
    port                = 80
    interval_in_seconds = 10
  }
}

resource "azurerm_traffic_manager_external_endpoint" "disabled_example" {
  name       = "example-endpoint"
  profile_id = azurerm_traffic_manager_profile.disabled_example.id
  weight     = 100
  target     = "sometarget"
}

resource "azurerm_resource_group" "germany_example" {
  name     = "example-resources"
  location = "Germany North"
}

resource "azurerm_traffic_manager_profile" "germany_example" {
  name                   = "example-profile"
  resource_group_name    = azurerm_resource_group.germany_example.name
  traffic_routing_method = "Weighted"

  dns_config {
    relative_name = "example-profile"
    ttl           = 100
  }

  monitor_config {
    protocol            = "HTTP"
    port                = 80
    interval_in_seconds = 10
  }
}

resource "azurerm_traffic_manager_external_endpoint" "germany_example" {
  name       = "example-endpoint"
  profile_id = azurerm_traffic_manager_profile.germany_example.id
  weight     = 100
  target     = "sometarget"
}
