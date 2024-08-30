provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

variable "fake_log_analytics_id" {
  type    = string
  default = "/subscriptions/12345678-1234-5678-90ab-1234567890ab/resourceGroups/MyRG/providers/Microsoft.OperationalInsights/workspaces/MyWS1"
}

resource "azurerm_monitor_data_collection_rule" "example" {
  for_each = toset(["with_usage", "without_usage"])

  name                = "example-rule-${each.key}"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location

  destinations {
    log_analytics {
      workspace_resource_id = var.fake_log_analytics_id
      name                  = "test-destination-log"
    }

    azure_monitor_metrics {
      name = "test-destination-metrics"
    }
  }

  data_flow {
    streams      = ["Microsoft-InsightsMetrics"]
    destinations = ["test-destination-metrics"]
  }

  data_sources {
    syslog {
      streams        = ["Microsoft-Syslog"]
      facility_names = ["*"]
      log_levels     = ["*"]
      name           = "test-datasource-syslog"
    }
  }
}

