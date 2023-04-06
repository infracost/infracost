provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

variable "fake_vm_1_id" {
  type    = string
  default = "/subscriptions/12345678-1234-5678-90ab-1234567890ab/resourceGroups/MyRG/providers/Microsoft.Compute/virtualMachines/MyVM1"
}

resource "azurerm_monitor_scheduled_query_rules_alert_v2" "example" {
  for_each = toset(["PT1M", "PT5M", "PT10M", "PT15M", "P1D"])

  name                = "example-${each.key}"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location

  evaluation_frequency = each.key
  window_duration      = each.key
  scopes               = [var.fake_vm_1_id]
  severity             = 4
  criteria {
    query                   = <<-QUERY
      requests
        | summarize CountByCountry=count() by client_CountryOrRegion
      QUERY
    time_aggregation_method = "Maximum"
    threshold               = 17.5
    operator                = "LessThan"

    resource_id_column    = "client_CountryOrRegion"
    metric_measure_column = "CountByCountry"
    dimension {
      name     = "client_CountryOrRegion"
      operator = "Exclude"
      values   = ["123"]
    }
  }

  enabled = true
}

resource "azurerm_monitor_scheduled_query_rules_alert_v2" "example-multi" {
  for_each = toset(["PT1M", "PT5M", "PT10M", "PT15M", "P1D"])

  name                = "example-multi-${each.key}"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location

  evaluation_frequency = each.key
  window_duration      = each.key
  scopes               = [var.fake_vm_1_id]
  severity             = 4
  criteria {
    query                   = <<-QUERY
      requests
        | summarize CountByCountry=count() by client_CountryOrRegion
      QUERY
    time_aggregation_method = "Maximum"
    threshold               = 17.5
    operator                = "LessThan"

    resource_id_column    = "client_CountryOrRegion"
    metric_measure_column = "CountByCountry"
    dimension {
      name     = "client_CountryOrRegion"
      operator = "Exclude"
      values   = ["123"]
    }

    dimension {
      name     = "client_OtherThing"
      operator = "Exclude"
      values   = ["123"]
    }
  }

  criteria {
    query                   = <<-QUERY
      requests
        | summarize CountByCountry=count() by client_CountryOrRegion
      QUERY
    time_aggregation_method = "Maximum"
    threshold               = 27.5
    operator                = "LessThan"

    resource_id_column    = "client_CountryOrRegion"
    metric_measure_column = "CountByCountry"
    dimension {
      name     = "client_CountryOrRegion"
      operator = "Exclude"
      values   = ["123"]
    }

    dimension {
      name     = "client_OtherThing"
      operator = "Exclude"
      values   = ["123"]
    }
  }

  enabled = true
}

resource "azurerm_monitor_scheduled_query_rules_alert_v2" "example-disabled" {
  name                = "example-disabled"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location

  evaluation_frequency = "PT10M"
  window_duration      = "PT10M"
  scopes               = [var.fake_vm_1_id]
  severity             = 4
  criteria {
    query                   = <<-QUERY
      requests
        | summarize CountByCountry=count() by client_CountryOrRegion
      QUERY
    time_aggregation_method = "Maximum"
    threshold               = 17.5
    operator                = "LessThan"

    resource_id_column    = "client_CountryOrRegion"
    metric_measure_column = "CountByCountry"
    dimension {
      name     = "client_CountryOrRegion"
      operator = "Exclude"
      values   = ["123"]
    }
  }

  enabled = false
}
