provider "azurerm" {
  skip_provider_registration = true
  features {}
}
resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "westus"
}
resource "azurerm_dns_zone" "westus" {
  name                = "mydomain.com"
  resource_group_name = azurerm_resource_group.example.name
}
resource "azurerm_resource_group" "example1" {
  name     = "example-resources"
  location = "germanywestcentral"
}
resource "azurerm_dns_zone" "germany" {
  name                = "mydomain.com"
  resource_group_name = azurerm_resource_group.example1.name
}
