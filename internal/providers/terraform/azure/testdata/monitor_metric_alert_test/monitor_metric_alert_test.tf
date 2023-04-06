provider "azurerm" {
  skip_provider_registration = true
  features {}
}

variable "fake_vm_1_id" {
  type    = string
  default = "/subscriptions/12345678-1234-5678-90ab-1234567890ab/resourceGroups/MyRG/providers/Microsoft.Compute/virtualMachines/MyVM1"
}

variable "fake_vm_2_id" {
  type    = string
  default = "/subscriptions/12345678-1234-5678-90ab-1234567890ab/resourceGroups/MyRG/providers/Microsoft.Compute/virtualMachines/MyVM2"
}


resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "eastus"
}

resource "azurerm_monitor_action_group" "example" {
  name                = "CriticalAlertsAction"
  resource_group_name = azurerm_resource_group.example.name
  short_name          = "p0action"
}

resource "azurerm_monitor_metric_alert" "example-single" {
  name                = "example-metricalert"
  resource_group_name = azurerm_resource_group.example.name
  scopes              = [var.fake_vm_1_id]

  enabled = true

  criteria {
    metric_namespace = "Microsoft.Storage/storageAccounts"
    metric_name      = "Transactions"
    aggregation      = "Total"
    operator         = "GreaterThan"
    threshold        = 50

    dimension {
      name     = "ApiName"
      operator = "Include"
      values   = ["*"]
    }
  }
}

resource "azurerm_monitor_metric_alert" "example-dynamic" {
  name                = "example-metricalert"
  resource_group_name = azurerm_resource_group.example.name
  scopes              = [var.fake_vm_1_id]

  dynamic_criteria {
    metric_namespace  = "Microsoft.Storage/storageAccounts"
    metric_name       = "Transactions"
    aggregation       = "Total"
    operator          = "GreaterThan"
    alert_sensitivity = "Low"

    dimension {
      name     = "ApiName"
      operator = "Include"
      values   = ["*"]
    }
  }
}

resource "azurerm_monitor_metric_alert" "example-multi" {
  name                = "example-multi"
  resource_group_name = azurerm_resource_group.example.name
  scopes              = [var.fake_vm_1_id, var.fake_vm_2_id]

  criteria {
    metric_namespace = "Microsoft.Storage/storageAccounts"
    metric_name      = "Transactions"
    aggregation      = "Total"
    operator         = "GreaterThan"
    threshold        = 50

    dimension {
      name     = "ApiName"
      operator = "Include"
      values   = ["X"]
    }

    dimension {
      name     = "ApiName"
      operator = "Include"
      values   = ["Y"]
    }
  }

  criteria {
    metric_namespace = "Microsoft.Storage/storageAccounts"
    metric_name      = "Transactions"
    aggregation      = "Total"
    operator         = "GreaterThan"
    threshold        = 60

    dimension {
      name     = "ApiName"
      operator = "Include"
      values   = ["X"]
    }

    dimension {
      name     = "ApiName"
      operator = "Include"
      values   = ["Y"]
    }
  }
}

resource "azurerm_monitor_metric_alert" "example-dynamic-multi" {
  name                = "example-dynamic-multi"
  resource_group_name = azurerm_resource_group.example.name
  scopes              = [var.fake_vm_1_id, var.fake_vm_2_id]

  dynamic_criteria {
    metric_namespace  = "Microsoft.Storage/storageAccounts"
    metric_name       = "Transactions"
    aggregation       = "Total"
    operator          = "GreaterThan"
    alert_sensitivity = "Low"

    dimension {
      name     = "ApiName"
      operator = "Include"
      values   = ["A"]
    }

    dimension {
      name     = "ApiName"
      operator = "Include"
      values   = ["B"]
    }
  }
}

resource "azurerm_monitor_metric_alert" "example-disabled" {
  name                = "example-disabled"
  resource_group_name = azurerm_resource_group.example.name
  scopes              = [var.fake_vm_1_id]

  enabled = false

  criteria {
    metric_namespace = "Microsoft.Storage/storageAccounts"
    metric_name      = "Transactions"
    aggregation      = "Total"
    operator         = "GreaterThan"
    threshold        = 50

    dimension {
      name     = "ApiName"
      operator = "Include"
      values   = ["*"]
    }
  }
}
