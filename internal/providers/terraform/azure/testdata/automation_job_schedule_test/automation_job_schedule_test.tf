provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "eastus"
}

resource "azurerm_automation_job_schedule" "zeroMinutes" {
  resource_group_name     = azurerm_resource_group.example.name
  automation_account_name = "tf-automation-account"
  schedule_name           = "hour"
  runbook_name            = "Get-VirtualMachine"
}
resource "azurerm_automation_job_schedule" "fiveMinutes" {
  resource_group_name     = azurerm_resource_group.example.name
  automation_account_name = "tf-automation-account"
  schedule_name           = "hour"
  runbook_name            = "Get-VirtualMachine"
}
resource "azurerm_automation_job_schedule" "withoutUsage" {
  resource_group_name     = azurerm_resource_group.example.name
  automation_account_name = "tf-automation-account"
  schedule_name           = "hour"
  runbook_name            = "Get-VirtualMachine"
}

