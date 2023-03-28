provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

locals {
  storage_modes                = ["GeoRedundant", "LocallyRedundant", "ZoneRedundant"]
  cross_region_restore_enabled = [true, false]
  skus                         = ["Standard", "RS0"]
  vault_permutations = distinct(flatten([
    for sku in local.skus : [
      for cross_region in local.cross_region_restore_enabled : [
        for storage_mode in local.storage_modes : {
          sku : sku
          storage_mode : storage_mode
          cross_region : cross_region
        }
      ]
    ]
  ]))
}

resource "azurerm_virtual_network" "example" {
  name                = "example-vnet"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
}

resource "azurerm_subnet" "example" {
  name                 = "internal"
  address_prefixes     = ["10.0.1.0/24"]
  resource_group_name  = azurerm_resource_group.example.name
  virtual_network_name = azurerm_virtual_network.example.name
}

resource "azurerm_network_interface" "example" {
  name                = "example-nic"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.example.id
    private_ip_address_allocation = "Dynamic"
  }
}


resource "azurerm_recovery_services_vault" "example" {
  for_each = { for entry in local.vault_permutations : "${entry.storage_mode}.${entry.sku}.${entry.cross_region}" => entry }

  name                         = "${each.value.storage_mode}-${each.value.sku}-${each.value.cross_region}"
  location                     = azurerm_resource_group.example.location
  resource_group_name          = azurerm_resource_group.example.name
  sku                          = each.value.sku
  storage_mode_type            = each.value.storage_mode
  cross_region_restore_enabled = each.value.cross_region
}

resource "azurerm_backup_policy_vm" "example" {
  for_each = { for entry in local.vault_permutations : "${entry.storage_mode}.${entry.sku}.${entry.cross_region}" => entry }

  name                = "policy-${each.value.storage_mode}-${each.value.sku}-${each.value.cross_region}"
  resource_group_name = azurerm_resource_group.example.name
  recovery_vault_name = "${each.value.storage_mode}-${each.value.sku}-${each.value.cross_region}"

  timezone = "UTC"

  backup {
    frequency = "Daily"
    time      = "23:00"
  }

  retention_daily {
    count = 10
  }
}

resource "azurerm_backup_protected_vm" "small" {
  for_each = { for entry in azurerm_backup_policy_vm.example : "${entry.name}" => entry }

  resource_group_name = azurerm_resource_group.example.name
  recovery_vault_name = each.value.recovery_vault_name
  source_vm_id        = azurerm_virtual_machine.small.id
  backup_policy_id    = each.value.id
}

resource "azurerm_backup_protected_vm" "medium" {
  for_each = { for entry in azurerm_backup_policy_vm.example : "${entry.name}" => entry }

  resource_group_name = azurerm_resource_group.example.name
  recovery_vault_name = each.value.recovery_vault_name
  source_vm_id        = azurerm_virtual_machine.medium.id
  backup_policy_id    = each.value.id
}

resource "azurerm_backup_protected_vm" "large" {
  for_each = { for entry in azurerm_backup_policy_vm.example : "${entry.name}" => entry }

  resource_group_name = azurerm_resource_group.example.name
  recovery_vault_name = each.value.recovery_vault_name
  source_vm_id        = azurerm_virtual_machine.large.id
  backup_policy_id    = each.value.id
}

resource "azurerm_backup_protected_vm" "with_usage" {
  for_each = { for entry in azurerm_backup_policy_vm.example : "${entry.name}" => entry }

  resource_group_name = azurerm_resource_group.example.name
  recovery_vault_name = each.value.recovery_vault_name
  source_vm_id        = azurerm_virtual_machine.small.id
  backup_policy_id    = each.value.id
}

resource "azurerm_virtual_machine" "small" {
  name                  = "small"
  location              = azurerm_resource_group.example.location
  resource_group_name   = azurerm_resource_group.example.name
  network_interface_ids = [azurerm_network_interface.example.id]
  vm_size               = "Standard_F2"

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  storage_os_disk {
    name              = "example-osdisk"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
    disk_size_gb      = 10
  }

  storage_data_disk {
    name              = "example-datadisk"
    lun               = 1
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
    disk_size_gb      = 30
  }

  os_profile {
    computer_name  = "examplevm"
    admin_username = "adminuser"
    admin_password = "P@ssword1234!"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }
}

resource "azurerm_virtual_machine" "medium" {
  name                  = "medium"
  location              = azurerm_resource_group.example.location
  resource_group_name   = azurerm_resource_group.example.name
  network_interface_ids = [azurerm_network_interface.example.id]
  vm_size               = "Standard_F2"

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  storage_os_disk {
    name              = "example-osdisk"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  storage_data_disk {
    name              = "example-datadisk"
    lun               = 1
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
    disk_size_gb      = 50
  }

  os_profile {
    computer_name  = "examplevm"
    admin_username = "adminuser"
    admin_password = "P@ssword1234!"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }
}

resource "azurerm_virtual_machine" "large" {
  name                  = "large"
  location              = azurerm_resource_group.example.location
  resource_group_name   = azurerm_resource_group.example.name
  network_interface_ids = [azurerm_network_interface.example.id]
  vm_size               = "Standard_F2"

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  storage_os_disk {
    name              = "example-osdisk"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  storage_data_disk {
    name              = "example-datadisk"
    lun               = 1
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
    disk_size_gb      = 500
  }

  storage_data_disk {
    name              = "example-datadisk-2"
    lun               = 1
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
    disk_size_gb      = 500
  }

  os_profile {
    computer_name  = "examplevm"
    admin_username = "adminuser"
    admin_password = "P@ssword1234!"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }
}
