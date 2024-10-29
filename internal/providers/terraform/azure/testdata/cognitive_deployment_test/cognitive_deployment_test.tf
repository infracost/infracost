provider "azurerm" {
  skip_provider_registration = true
  features {}
}

locals {
  models = [
    "gpt-4",
    "gpt-4-32k",
    "gpt-35-turbo",
    "gpt-35-turbo-16k",
    "gpt-35-turbo-instruct",
    "text-embedding-ada-002",
    "text-embedding-3-small",
    "text-embedding-3-large",
    "babbage-002",
    "dall-e-2",
    "dall-e-3",
    "davinci-002",
    "tts",
    "tts-hd",
    "whisper",
  ]

  versions = {
    "gpt-4" : ["1106-preview", "0125-preview", "vision-preview"]
  }

  permutations = flatten([
    for model in local.models : [
      concat([
        {
          key     = model
          model   = model
          version = null
        }
        ], [
        for v in lookup(local.versions, model, []) : {
          key     = "${model}-${v}"
          model   = model
          version = v
        }
      ])
    ]
  ])
}

resource "azurerm_resource_group" "eastus" {
  name     = "eastus"
  location = "eastus"
}

resource "azurerm_resource_group" "eastus2" {
  name     = "eastus2"
  location = "eastus2"
}

resource "azurerm_resource_group" "swedencentral" {
  name     = "swedencentral"
  location = "swedencentral"
}

resource "azurerm_cognitive_account" "eastus" {
  name                = "eastus"
  resource_group_name = azurerm_resource_group.eastus.name
  location            = azurerm_resource_group.eastus.location
  kind                = "OpenAI"
  sku_name            = "S0"
}

resource "azurerm_cognitive_account" "eastus2" {
  name                = "eastus2"
  resource_group_name = azurerm_resource_group.eastus2.name
  location            = azurerm_resource_group.eastus2.location
  kind                = "OpenAI"
  sku_name            = "S0"
}

resource "azurerm_cognitive_account" "swedencentral" {
  name                = "swedencentral"
  resource_group_name = azurerm_resource_group.swedencentral.name
  location            = azurerm_resource_group.swedencentral.location
  kind                = "OpenAI"
  sku_name            = "S0"
}

resource "azurerm_cognitive_deployment" "eastus_without_usage" {
  for_each = { for p in local.permutations : "${p.key}" => p }

  name                 = "eastus-without-usage-${each.key}"
  cognitive_account_id = azurerm_cognitive_account.eastus.id

  model {
    format  = "OpenAI"
    name    = each.value.model
    version = each.value.version
  }

  sku {
    name = "Standard"
  }
}

resource "azurerm_cognitive_deployment" "eastus2_without_usage" {
  for_each = { for p in local.permutations : "${p.key}" => p }

  name                 = "eastus2-without-usage-${each.key}"
  cognitive_account_id = azurerm_cognitive_account.eastus2.id

  model {
    format  = "OpenAI"
    name    = each.value.model
    version = each.value.version
  }

  sku {
    name = "Standard"
  }
}

resource "azurerm_cognitive_deployment" "swedencentral_without_usage" {
  for_each = { for p in local.permutations : "${p.key}" => p }

  name                 = "swedencentral-without-usage-${each.key}"
  cognitive_account_id = azurerm_cognitive_account.swedencentral.id

  model {
    format  = "OpenAI"
    name    = each.value.model
    version = each.value.version
  }

  sku {
    name = "Standard"
  }
}

resource "azurerm_cognitive_deployment" "eastus_with_usage" {
  for_each = { for p in local.permutations : "${p.key}" => p }

  name                 = "eastus-with-usage-${each.key}"
  cognitive_account_id = azurerm_cognitive_account.eastus.id

  model {
    format  = "OpenAI"
    name    = each.value.model
    version = each.value.version
  }

  sku {
    name = "Standard"
  }
}

resource "azurerm_cognitive_deployment" "eastus2_with_usage" {
  for_each = { for p in local.permutations : "${p.key}" => p }

  name                 = "eastus2-with-usage-${each.key}"
  cognitive_account_id = azurerm_cognitive_account.eastus2.id

  model {
    format  = "OpenAI"
    name    = each.value.model
    version = each.value.version
  }

  sku {
    name = "Standard"
  }
}

resource "azurerm_cognitive_deployment" "swedencentral_with_usage" {
  for_each = { for p in local.permutations : "${p.key}" => p }

  name                 = "swedencentral-with-usage-${each.key}"
  cognitive_account_id = azurerm_cognitive_account.swedencentral.id

  model {
    format  = "OpenAI"
    name    = each.value.model
    version = each.value.version
  }

  sku {
    name = "Standard"
  }
}

resource "azurerm_cognitive_deployment" "free_tier" {
  name                 = "free-tier"
  cognitive_account_id = azurerm_cognitive_account.eastus.id

  model {
    format = "OpenAI"
    name   = "gpt-4"
  }

  sku {
    name = "Standard"
  }
}

resource "azurerm_cognitive_deployment" "unsupported" {
  name                 = "unsupported"
  cognitive_account_id = azurerm_cognitive_account.eastus.id

  model {
    format = "OpenAI"
    name   = "ada"
  }

  sku {
    name = "Standard"
  }
}
