provider "azurerm" {
  skip_provider_registration = true
  features {}
}

locals {
  vm_sizes = [
    "Basic_A4",
    "Standard_A8m_v2",
    "Standard_A9",
    "Standard_B8as_v2",
    "Standard_B8ms",
    "Standard_B8ps_v2",
    "Standard_B8s_v2",
    "Standard_D4",
    "Standard_D5_v2",
    "Standard_D8_v3",
    "Standard_D8_v4",
    "Standard_D8d_v4",
    "Standard_D8ds_v4",
    "Standard_D8pds_v5",
    "Standard_D8plds_v5",
    "Standard_D8pls_v5",
    "Standard_D8ps_v5",
    "Standard_D8s_v3",
    "Standard_D8s_v4",
    "Standard_D96_v5",
    "Standard_D96a_v4",
    "Standard_D96ads_v5",
    "Standard_D96ads_v6",
    "Standard_D96alds_v6",
    "Standard_D96als_v6",
    "Standard_D96as_v4",
    "Standard_D96as_v5",
    "Standard_D96as_v6",
    "Standard_D96d_v5",
    "Standard_D96ds_v5",
    "Standard_D96ds_v6",
    "Standard_D96lds_v5",
    "Standard_D96lds_v6",
    "Standard_D96ls_v5",
    "Standard_D96ls_v6",
    "Standard_D96s_v5",
    "Standard_D96s_v6",
    "Standard_DC4s",
    "Standard_DC8_v2",
    "Standard_DC8ds_v3",
    "Standard_DC8s_v3",
    "Standard_DC96ads_cc_v5",
    "Standard_DC96ads_v5",
    "Standard_DC96as_cc_v5",
    "Standard_DC96as_v5",
    "Standard_DS4",
    "Standard_DS5_v2",
    "Standard_E8_v3",
    "Standard_E8_v4",
    "Standard_E8d_v4",
    "Standard_E8ds_v4",
    "Standard_E8pds_v5",
    "Standard_E8ps_v5",
    "Standard_E8s_v3",
    "Standard_E8s_v4",
    "Standard_E96_v5",
    "Standard_E96a_v4",
    "Standard_E96ads_v5",
    "Standard_E96ads_v6",
    "Standard_E96as_v4",
    "Standard_E96as_v5",
    "Standard_E96as_v6",
    "Standard_E96bds_v5",
    "Standard_E96bs_v5",
    "Standard_E96d_v5",
    "Standard_E96ds_v5",
    "Standard_E96ds_v6",
    "Standard_E96iads_v5",
    "Standard_E96ias_v5",
    "Standard_E96s_v5",
    "Standard_E96s_v6",
    "Standard_EC96ads_cc_v5",
    "Standard_EC96as_cc_v5",
    "Standard_EC96iads_v5",
    "Standard_EC96ias_v5",
    "Standard_F8",
    "Standard_F8ads_v6",
    "Standard_F8alds_v6",
    "Standard_F8als_v6",
    "Standard_F8amds_v6",
    "Standard_F8ams_v6",
    "Standard_F8as_v6",
    "Standard_F8s",
    "Standard_F8s_v2",
    "Standard_FX4mds",
    "Standard_H8m",
    "Standard_HB120rs_v2",
    "Standard_HB120rs_v3",
    "Standard_HB176rs_v4",
    "Standard_HB60rs",
    "Standard_HC44rs",
    "Standard_HX176rs",
    "Standard_L8as_v3",
    "Standard_L8s_v2",
    "Standard_L8s_v3",
    "Standard_M64ds_v2",
    "Standard_M64s_v2",
    "Standard_M8ms",
    "Standard_M96ds_2_v3",
    "Standard_M96s_2_v3",
    "Standard_NC6",
    "Standard_NC6s_v2",
    "Standard_NC6s_v3",
    "Standard_NC80adis_H100_v5",
    "Standard_NC8as_T4_v3",
    "Standard_NC96ads_A100_v4",
    "Standard_ND40rs_v2",
    "Standard_ND40s_v2",
    "Standard_ND6s",
    "Standard_ND96amsr_A100_v4",
    "Standard_ND96asr_v4",
    "Standard_ND96isr_H100_v5",
    "Standard_NP40s",
    "Standard_NV48s_v3",
    "Standard_NV6",
    "Standard_NV72ads_A10_v5",
    "Standard_NV8as_v4",
    "Standard_PB6s",
  ]
}


resource "azurerm_resource_group" "main" {
  name     = "example-resources"
  location = "eastus"
}

resource "azurerm_virtual_network" "main" {
  name                = "network"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
}

resource "azurerm_subnet" "internal" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.main.name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = ["10.0.2.0/24"]
}

resource "azurerm_network_interface" "main" {
  name                = "-nic"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = azurerm_subnet.internal.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_virtual_machine" "linux" {
  name                  = "vm"
  location              = azurerm_resource_group.main.location
  resource_group_name   = azurerm_resource_group.main.name
  network_interface_ids = [azurerm_network_interface.main.id]
  vm_size               = "Standard_DS1_v2"

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
  storage_os_disk {
    name              = "myosdisk1"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }
  os_profile {
    computer_name  = "hostname"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }
  os_profile_linux_config {
    disable_password_authentication = false
  }
}

resource "azurerm_virtual_machine" "linux_withMonthlyHours" {
  name                  = "vm"
  location              = azurerm_resource_group.main.location
  resource_group_name   = azurerm_resource_group.main.name
  network_interface_ids = [azurerm_network_interface.main.id]
  vm_size               = "Standard_DS1_v2"

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }
  storage_os_disk {
    name              = "myosdisk1"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }
  os_profile {
    computer_name  = "hostname"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }
  os_profile_linux_config {
    disable_password_authentication = false
  }
}

resource "azurerm_virtual_machine" "windows" {
  name                  = "vm"
  location              = azurerm_resource_group.main.location
  resource_group_name   = azurerm_resource_group.main.name
  network_interface_ids = [azurerm_network_interface.main.id]
  vm_size               = "Standard_DS1_v2"

  storage_os_disk {
    name              = "myosdisk1"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
    os_type           = "Windows"
  }
  os_profile {
    computer_name  = "hostname"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }
  storage_data_disk {
    name              = "myosdisk1"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
    lun               = 1
  }
  storage_data_disk {
    name              = "myosdisk1"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "StandardSSD_LRS"
    lun               = 2
  }
  storage_data_disk {
    name              = "myosdisk1"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Premium_LRS"
    lun               = 3
  }
  storage_data_disk {
    name              = "myosdisk1"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "UltraSSD_LRS"
    lun               = 4
  }
}

resource "azurerm_virtual_machine" "windows_withMonthlyHours" {
  name                  = "vm"
  location              = azurerm_resource_group.main.location
  resource_group_name   = azurerm_resource_group.main.name
  network_interface_ids = [azurerm_network_interface.main.id]
  vm_size               = "Standard_DS1_v2"

  storage_os_disk {
    name              = "myosdisk1"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
    os_type           = "Windows"
  }
  os_profile {
    computer_name  = "hostname"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }
}

resource "azurerm_virtual_machine" "linux_vms" {
  for_each = toset(local.vm_sizes)

  name                  = "vm"
  location              = azurerm_resource_group.main.location
  resource_group_name   = azurerm_resource_group.main.name
  network_interface_ids = [azurerm_network_interface.main.id]
  vm_size               = each.value

  storage_os_disk {
    name              = "myosdisk1"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }
}

resource "azurerm_virtual_machine" "windows_vms" {
  for_each = toset(local.vm_sizes)

  name                  = "vm"
  location              = azurerm_resource_group.main.location
  resource_group_name   = azurerm_resource_group.main.name
  network_interface_ids = [azurerm_network_interface.main.id]
  vm_size               = each.value

  storage_os_disk {
    name              = "myosdisk1"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
    os_type           = "Windows"
  }
}

