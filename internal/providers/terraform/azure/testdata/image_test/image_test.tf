provider "azurerm" {
  features {}
  skip_provider_registration = true
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_virtual_network" "example" {
  name                = "example"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
}

resource "azurerm_subnet" "example" {
  name                 = "example"
  address_prefixes     = ["10.0.1.0/24"]
  resource_group_name  = azurerm_resource_group.example.name
  virtual_network_name = azurerm_virtual_network.example.name
}

resource "azurerm_network_interface" "example" {
  name                = "test-nic"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = azurerm_subnet.example.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_managed_disk" "example" {
  name                 = "examplemd"
  location             = azurerm_resource_group.example.location
  resource_group_name  = azurerm_resource_group.example.name
  storage_account_type = "Standard_LRS"
  create_option        = "Empty"
  disk_size_gb         = 70
}

locals {
  permuations = [
    {
      name            = "standard"
      data_disk       = []
      disk_type       = "StandardSSD_LRS"
      vm_size         = "Standard_D1_v2"
      managed_disk_id = null
    },
    {
      name            = "premium"
      data_disk       = []
      disk_type       = "Premium_LRS"
      vm_size         = "Standard_DS1_v2"
      managed_disk_id = null
    },
    {
      name            = "md"
      data_disk       = []
      disk_type       = "Premium_LRS"
      vm_size         = "Standard_DS1_v2"
      managed_disk_id = azurerm_managed_disk.example.id
    },
    {
      name            = "data_disk"
      disk_type       = "Premium_LRS"
      vm_size         = "Standard_DS1_v2"
      managed_disk_id = null
      data_disk = [
        {
          disk_size_gb      = 30
          managed_disk_type = "Premium_LRS"
          create_option     = "Empty"
          lun               = "10"
          name              = "premiumdata"
          managed_disk_id   = null
        },
        {
          disk_size_gb      = null
          managed_disk_type = "Premium_LRS"
          create_option     = "Empty"
          lun               = "10"
          name              = "premiumdata"
          managed_disk_id   = azurerm_managed_disk.example.id
        }
      ]
    }
  ]
}

resource "azurerm_virtual_machine" "example" {
  for_each = { for entry in local.permuations : "${entry.name}" => entry }

  name                  = each.value.name
  location              = azurerm_resource_group.example.location
  resource_group_name   = azurerm_resource_group.example.name
  network_interface_ids = [azurerm_network_interface.example.id]
  vm_size               = each.value.vm_size

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
    managed_disk_type = each.value.disk_type
  }

  dynamic "storage_data_disk" {
    for_each = each.value.data_disk

    content {
      disk_size_gb      = storage_data_disk.value.disk_size_gb
      managed_disk_type = storage_data_disk.value.managed_disk_type
      create_option     = storage_data_disk.value.create_option
      lun               = storage_data_disk.value.lun
      name              = storage_data_disk.value.name
    }
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

resource "azurerm_image" "vm" {
  for_each = { for entry in azurerm_virtual_machine.example : "${entry.name}" => entry }

  name                      = each.value.name
  location                  = azurerm_resource_group.example.location
  resource_group_name       = azurerm_resource_group.example.name
  source_virtual_machine_id = each.value.id
}

resource "azurerm_image" "vm_usage" {
  for_each = { for entry in azurerm_virtual_machine.example : "${entry.name}" => entry }

  name                      = each.value.name
  location                  = azurerm_resource_group.example.location
  resource_group_name       = azurerm_resource_group.example.name
  source_virtual_machine_id = each.value.id
}

locals {
  disk_permutations = [
    {
      name            = "spec"
      size_gb         = 30
      data_disk       = []
      managed_disk_id = null
    },
    {
      name            = "data"
      size_gb         = 30
      managed_disk_id = null
      data_disk = [
        {
          size_gb         = 30
          managed_disk_id = null
        }
      ]
    },
    {
      name            = "managed"
      size_gb         = null
      managed_disk_id = azurerm_managed_disk.example.id
      data_disk = [
        {
          size_gb         = null
          managed_disk_id = azurerm_managed_disk.example.id
        }
      ]
    }
  ]
}

resource "azurerm_image" "disk" {
  for_each = { for entry in local.disk_permutations : "${entry.name}" => entry }

  name                = each.value.name
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name

  os_disk {
    size_gb         = each.value.size_gb
    managed_disk_id = each.value.managed_disk_id
  }

  dynamic "data_disk" {
    for_each = each.value.data_disk
    content {
      size_gb         = data_disk.value.size_gb
      managed_disk_id = data_disk.value.managed_disk_id
    }
  }
}
