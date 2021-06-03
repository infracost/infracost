provider "azurerm" {
  skip_provider_registration = true
  features {}
   subscription_id             = "84171726-c002-4fc7-924a-b6ff82c59677"
  client_id                   = "eec7409b-6cbd-420d-b3ba-bc0ef14c4bbf"
  client_secret               = "OZ~wK2ZpptK2O1Z61_qD2ycu~1KQ4WDzvL"
  tenant_id                   = "b0b4f93e-59ae-44ff-b39a-e51e7e2b16e9"
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "East US"
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