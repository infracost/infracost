provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_windows_virtual_machine_scale_set" "basic_a2" {
  name                = "basic_a2"
  resource_group_name = "fake_resource_group"
  location            = "eastus"
  instances           = 3

  sku            = "Basic_A2"
  admin_username = "fakeuser"
  admin_password = "Password1234!"

  network_interface {
    name    = "example"
    primary = true

    ip_configuration {
      name      = "internal"
      primary   = true
      subnet_id = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Network/virtualNetworks/test1/subnets/fakesubnet"
    }
  }

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "MicrosoftWindowsServer"
    offer     = "WindowsServer"
    sku       = "2016-Datacenter"
    version   = "latest"
  }
}

resource "azurerm_windows_virtual_machine_scale_set" "basic_a2_usage" {
  name                = "basic_a2"
  resource_group_name = "fake_resource_group"
  location            = "eastus"
  instances           = 3

  sku            = "Basic_A2"
  admin_username = "fakeuser"
  admin_password = "Password1234!"

  network_interface {
    name    = "example"
    primary = true

    ip_configuration {
      name      = "internal"
      primary   = true
      subnet_id = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Network/virtualNetworks/test1/subnets/fakesubnet"
    }
  }

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "MicrosoftWindowsServer"
    offer     = "WindowsServer"
    sku       = "2016-Datacenter"
    version   = "latest"
  }
}
