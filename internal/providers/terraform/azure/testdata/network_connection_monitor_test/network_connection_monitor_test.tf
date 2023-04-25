provider "azurerm" {
  skip_provider_registration = true
  features {}
}

locals {
  permutations = {
    "single-test" = {
      test_groups = [
        {
          name                  = "test-group-1"
          destination_endpoints = ["destination-1"]
          source_endpoints      = ["source-1"]
          test_configurations   = ["tcp-test-1"]
        }
      ]
    }
    "12-tests" = {
      test_groups = [
        {
          name                  = "test-group-1"
          destination_endpoints = ["destination-1", "destination-2"]
          source_endpoints      = ["source-1", "source-2"]
          test_configurations   = ["tcp-test-1", "tcp-test-2", "tcp-test-3"]
        }
      ]
    }
    "multiple-test-groups-18-tests" = {
      test_groups = [
        {
          name                  = "test-group-1"
          destination_endpoints = ["destination-1", "destination-2"]
          source_endpoints      = ["source-1", "source-2"]
          test_configurations   = ["tcp-test-1", "tcp-test-2", "tcp-test-3"]
        },
        {
          name                  = "test-group-2"
          destination_endpoints = ["destination-1"]
          source_endpoints      = ["source-1", "source-2"]
          test_configurations   = ["tcp-test-1", "tcp-test-2", "tcp-test-3"]
        }
      ]
    }
    "multiple-test-groups-1-disabled" = {
      test_groups = [
        {
          name                  = "test-group-1"
          destination_endpoints = ["destination-1", "destination-2"]
          source_endpoints      = ["source-1", "source-2"]
          test_configurations   = ["tcp-test-1", "tcp-test-2", "tcp-test-3"]
        },
        {
          name                  = "test-group-2"
          enabled               = false
          destination_endpoints = ["destination-1", "destination-2"]
          source_endpoints      = ["source-1", "source-2"]
          test_configurations   = ["tcp-test-1", "tcp-test-2", "tcp-test-3"]
        }
      ]
    }
  }
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_network_watcher" "network_watcher" {
  name                = "example-Watcher"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
}

resource "azurerm_network_connection_monitor" "connection_monitor" {
  for_each = { for key, value in local.permutations : key => value }

  name               = "example-monitor"
  network_watcher_id = azurerm_network_watcher.network_watcher.id
  location           = azurerm_network_watcher.network_watcher.location

  endpoint {
    name    = "source"
    address = "example.com"
  }

  endpoint {
    name    = "destination"
    address = "example.org"
  }

  test_configuration {
    name                      = "tcp-test"
    protocol                  = "Tcp"
    test_frequency_in_seconds = 60

    tcp_configuration {
      port = 80
    }
  }

  dynamic "test_group" {
    for_each = each.value.test_groups
    content {
      name                     = test_group.value.name
      enabled                  = lookup(test_group.value, "enabled", true)
      destination_endpoints    = test_group.value.destination_endpoints
      source_endpoints         = test_group.value.source_endpoints
      test_configuration_names = test_group.value.test_configurations
    }
  }
}

resource "azurerm_network_connection_monitor" "connection_monitor_with_usage" {
  name               = "example-monitor"
  network_watcher_id = azurerm_network_watcher.network_watcher.id
  location           = azurerm_network_watcher.network_watcher.location

  endpoint {
    name    = "source"
    address = "example.com"
  }

  endpoint {
    name    = "destination"
    address = "example.org"
  }

  test_configuration {
    name                      = "tcp-test"
    protocol                  = "Tcp"
    test_frequency_in_seconds = 60

    tcp_configuration {
      port = 80
    }
  }

  test_group {
    name                     = "test-group-1"
    destination_endpoints    = ["destination-1", "destination-2"]
    source_endpoints         = ["source-1", "source-2"]
    test_configuration_names = ["tcp-test-1", "tcp-test-2", "tcp-test-3"]
  }
}
