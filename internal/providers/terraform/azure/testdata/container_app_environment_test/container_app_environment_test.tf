provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_log_analytics_workspace" "example" {
  name                = "example-law"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "PerGB2018"
}

resource "azurerm_container_app_environment" "example_dedicated" {
  name                       = "example-env-dedicated"
  location                   = azurerm_resource_group.example.location
  resource_group_name        = azurerm_resource_group.example.name
  log_analytics_workspace_id = azurerm_log_analytics_workspace.example.id

  workload_profile {
    name                  = "Consumption"
    workload_profile_type = "Consumption"
  }

  workload_profile {
    name                  = "MyDedicatedProfile"
    workload_profile_type = "D4"
    minimum_count         = 1
    maximum_count         = 3
  }

  workload_profile {
    name                  = "MyMemoryOptimizedProfile"
    workload_profile_type = "E4"
    minimum_count         = 1
    maximum_count         = 3
  }
}

resource "azurerm_container_app" "example" {
  name                         = "example-app"
  container_app_environment_id = azurerm_container_app_environment.example_dedicated.id
  resource_group_name          = azurerm_resource_group.example.name
  revision_mode                = "Single"
  workload_profile_name        = "MyDedicatedProfile"

  template {
    container {
      name   = "examplecontainerapp"
      image  = "mcr.microsoft.com/azuredocs/containerapps-helloworld:latest"
      cpu    = 0.5
      memory = "1Gi"
    }
  }
}
