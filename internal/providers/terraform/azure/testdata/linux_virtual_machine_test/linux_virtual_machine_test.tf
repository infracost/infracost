provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_linux_virtual_machine" "basic_b1" {
  name                = "basic_b1"
  resource_group_name = "fake_resource_group"
  location            = "eastus"

  size           = "Standard_B1s"
  admin_username = "fakeuser"
  admin_password = "Password1234!"

  network_interface_ids = [
    "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Network/networkInterfaces/fakenic",
  ]

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}

resource "azurerm_linux_virtual_machine" "basic_b1_lowercase" {
  name                = "basic_b1"
  resource_group_name = "fake_resource_group"
  location            = "eastus"

  size           = "standard_b1s"
  admin_username = "fakeuser"
  admin_password = "Password1234!"

  network_interface_ids = [
    "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Network/networkInterfaces/fakenic",
  ]

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}

resource "azurerm_linux_virtual_machine" "basic_a2" {
  name                = "basic_a2"
  resource_group_name = "fake_resource_group"
  location            = "eastus"

  size           = "Basic_A2"
  admin_username = "fakeuser"
  admin_password = "Password1234!"

  network_interface_ids = [
    "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Network/networkInterfaces/fakenic",
  ]

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}

resource "azurerm_linux_virtual_machine" "standard_f2_premium_disk" {
  name                = "standard_f2"
  resource_group_name = "fake_resource_group"
  location            = "eastus"

  size           = "Standard_F2"
  admin_username = "fakeuser"
  admin_password = "Password1234!"

  network_interface_ids = [
    "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Network/networkInterfaces/fakenic",
  ]

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Premium_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}

resource "azurerm_linux_virtual_machine" "standard_f2_lowercase" {
  name                = "standard_f2"
  resource_group_name = "fake_resource_group"
  location            = "eastus"

  size           = "standard_f2"
  admin_username = "fakeuser"
  admin_password = "Password1234!"

  network_interface_ids = [
    "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Network/networkInterfaces/fakenic",
  ]

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Premium_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}

resource "azurerm_linux_virtual_machine" "standard_a2_v2_custom_disk" {
  name                = "standard_a2_v2_custom_disk"
  resource_group_name = "fake_resource_group"
  location            = "eastus"

  size           = "Standard_A2_v2"
  admin_username = "fakeuser"
  admin_password = "Password1234!"

  network_interface_ids = [
    "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Network/networkInterfaces/fakenic",
  ]

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "StandardSSD_LRS"
    disk_size_gb         = 1000
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}

resource "azurerm_linux_virtual_machine" "standard_a2_ultra_enabled" {
  name                = "standard_a2_ultra_enabled"
  resource_group_name = "fake_resource_group"
  location            = "eastus"

  size           = "Standard_A2_v2"
  admin_username = "fakeuser"
  admin_password = "Password1234!"

  network_interface_ids = [
    "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Network/networkInterfaces/fakenic",
  ]

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "StandardSSD_LRS"
  }

  additional_capabilities {
    ultra_ssd_enabled = true
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}

resource "azurerm_linux_virtual_machine" "basic_b1_withMonthlyHours" {
  name                = "basic_b1"
  resource_group_name = "fake_resource_group"
  location            = "eastus"

  size           = "Standard_B1s"
  admin_username = "fakeuser"
  admin_password = "Password1234!"

  network_interface_ids = [
    "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Network/networkInterfaces/fakenic",
  ]

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
}
