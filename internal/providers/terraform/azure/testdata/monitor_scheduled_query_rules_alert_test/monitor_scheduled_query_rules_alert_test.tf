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

resource "azurerm_monitor_scheduled_query_rules_alert" "example_freq_5" {
  for_each            = toset(["5", "11", "15", "60"])
  name                = "example-${each.key}"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name

  action {
    action_group           = []
    email_subject          = "Email Header"
    custom_webhook_payload = "{}"
  }

  data_source_id = var.fake_vm_1_id
  enabled        = true
  # Count all requests with server error result code grouped into 5-minute bins
  query       = <<-QUERY
  requests
    | where tolong(resultCode) >= 500
    | summarize count() by bin(timestamp, 5m)
  QUERY
  severity    = 1
  frequency   = each.key
  time_window = 30
  trigger {
    operator  = "GreaterThan"
    threshold = 3
  }
}

#resource "azurerm_monitor_scheduled_query_rules_alert" "example_freq_10" {
#  name                = "example"
#  location            = azurerm_resource_group.example.location
#  resource_group_name = azurerm_resource_group.example.name
#
#  action {
#    action_group           = []
#    email_subject          = "Email Header"
#    custom_webhook_payload = "{}"
#  }
#
#  data_source_id = var.fake_vm_1_id
#  enabled        = true
#  # Count all requests with server error result code grouped into 5-minute bins
#  query       = <<-QUERY
#  requests
#    | where tolong(resultCode) >= 500
#    | summarize count() by bin(timestamp, 5m)
#  QUERY
#  severity    = 1
#  frequency   = 10
#  time_window = 30
#  trigger {
#    operator  = "GreaterThan"
#    threshold = 3
#  }
#}

#resource "azurerm_monitor_scheduled_query_rules_alert" "example_freq_60" {
#  name                = "example"
#  location            = azurerm_resource_group.example.location
#  resource_group_name = azurerm_resource_group.example.name
#
#  action {
#    action_group           = []
#    email_subject          = "Email Header"
#    custom_webhook_payload = "{}"
#  }
#
#  data_source_id = var.fake_vm_1_id
#  enabled        = true
#  # Count all requests with server error result code grouped into 5-minute bins
#  query       = <<-QUERY
#  requests
#    | where tolong(resultCode) >= 500
#    | summarize count() by bin(timestamp, 5m)
#  QUERY
#  severity    = 1
#  frequency   = 60
#  time_window = 30
#  trigger {
#    operator  = "GreaterThan"
#    threshold = 3
#  }
#}

resource "azurerm_monitor_scheduled_query_rules_alert" "example_disabled" {
  name                = "example"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name

  action {
    action_group           = []
    email_subject          = "Email Header"
    custom_webhook_payload = "{}"
  }

  data_source_id = var.fake_vm_1_id
  enabled        = false
  # Count all requests with server error result code grouped into 5-minute bins
  query       = <<-QUERY
  requests
    | where tolong(resultCode) >= 500
    | summarize count() by bin(timestamp, 5m)
  QUERY
  severity    = 1
  frequency   = 60
  time_window = 30
  trigger {
    operator  = "GreaterThan"
    threshold = 3
  }
}
