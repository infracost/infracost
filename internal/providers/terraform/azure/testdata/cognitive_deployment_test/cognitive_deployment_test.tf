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
    "gpt-4o",
    "gpt-4o-mini",
    "gpt-4.5-preview",
    "gpt-4o-realtime-preview",
    "gpt-4o-mini-realtime-preview",
    "gpt-4o-audio-preview",
    "gpt-4o-mini-audio-preview",
    "computer-use-preview",
    "o1",
    "o1-mini",
    "o3-mini",
  ]

  versions = {
    "gpt-4" : ["1106-preview", "0125-preview"],
    "gpt-4o" : ["2024-05-13", "2024-08-06"],
    "gpt-4o-mini" : ["2024-07-18"],
    "gpt-4.5-preview" : ["2025-02-27"],
    "gpt-4o-realtime-preview" : ["2024-12-17", "2024-10-01"],
    "gpt-4o-mini-realtime-preview" : ["2024-12-17"],
    "gpt-4o-audio-preview" : ["2024-12-17"],
    "gpt-4o-mini-audio-preview" : ["2024-12-17"],
    "computer-use-preview" : ["global"],
    "o1" : ["2024-12-17"],
    "o1-mini" : ["2024-09-12"],
    "o3-mini" : ["2025-01-31"],
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

# New specific test resources for audio, realtime, and file search functionality
resource "azurerm_cognitive_deployment" "audio_models" {
  for_each = {
    "gpt4o_audio" = {
      model   = "gpt-4o-audio-preview"
      version = "2024-12-17"
    }
    "gpt4o_mini_audio" = {
      model   = "gpt-4o-mini-audio-preview"
      version = "2024-12-17"
    }
  }

  name                 = "audio-model-${each.key}"
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

resource "azurerm_cognitive_deployment" "realtime_models" {
  for_each = {
    "gpt4o_realtime_1217" = {
      model   = "gpt-4o-realtime-preview"
      version = "2024-12-17"
    }
    "gpt4o_realtime_1001" = {
      model   = "gpt-4o-realtime-preview"
      version = "2024-10-01"
    }
    "gpt4o_mini_realtime" = {
      model   = "gpt-4o-mini-realtime-preview"
      version = "2024-12-17"
    }
  }

  name                 = "realtime-model-${each.key}"
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

resource "azurerm_cognitive_deployment" "computer_use_models" {
  name                 = "computer-use-preview"
  cognitive_account_id = azurerm_cognitive_account.eastus.id

  model {
    format  = "OpenAI"
    name    = "computer-use-preview"
    version = "global"
  }

  sku {
    name = "Standard"
  }
}

resource "azurerm_cognitive_deployment" "new_foundation_models" {
  for_each = {
    "gpt45_preview" = {
      model   = "gpt-4.5-preview"
      version = "2025-02-27"
    }
    "o1_model" = {
      model   = "o1"
      version = "2024-12-17"
    }
    "o1_mini_model" = {
      model   = "o1-mini"
      version = "2024-09-12"
    }
    "o3_mini_model" = {
      model   = "o3-mini"
      version = "2025-01-31"
    }
  }

  name                 = "foundation-model-${each.key}"
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

# Provisioned throughput unit tests
resource "azurerm_cognitive_deployment" "provisioned_regional" {
  for_each = {
    "gpt4o" = {
      model    = "gpt-4o"
      version  = "2024-05-13"
      capacity = 1
    }
    "gpt4o_high_capacity" = {
      model    = "gpt-4o"
      version  = "2024-05-13"
      capacity = 5
    }
    "gpt35_turbo" = {
      model    = "gpt-35-turbo"
      version  = "0125"
      capacity = 2
    }
  }

  name                 = "provisioned-regional-${each.key}"
  cognitive_account_id = azurerm_cognitive_account.eastus.id

  model {
    format  = "OpenAI"
    name    = each.value.model
    version = each.value.version
  }

  sku {
    name     = "ProvisionedManaged"
    capacity = each.value.capacity
  }
}

resource "azurerm_cognitive_deployment" "provisioned_global" {
  for_each = {
    "gpt4" = {
      model    = "gpt-4"
      version  = "0613"
      capacity = 1
    }
    "gpt4_32k" = {
      model    = "gpt-4-32k"
      version  = "0613"
      capacity = 3
    }
  }

  name                 = "provisioned-global-${each.key}"
  cognitive_account_id = azurerm_cognitive_account.eastus.id

  model {
    format  = "OpenAI"
    name    = each.value.model
    version = each.value.version
  }

  sku {
    name     = "GlobalProvisionedManaged"
    capacity = each.value.capacity
  }
}

resource "azurerm_cognitive_deployment" "provisioned_data_zone" {
  for_each = {
    "gpt4o_data_zone" = {
      model    = "gpt-4o"
      version  = "2024-05-13"
      capacity = 2
    }
    "o1_mini_data_zone" = {
      model    = "o1-mini"
      version  = "2024-09-12"
      capacity = 4
    }
  }

  name                 = "provisioned-data-zone-${each.key}"
  cognitive_account_id = azurerm_cognitive_account.eastus.id

  model {
    format  = "OpenAI"
    name    = each.value.model
    version = each.value.version
  }

  sku {
    name     = "DataZoneProvisionedManaged"
    capacity = each.value.capacity
  }
}

# Regional provisioned unit tests for different regions
resource "azurerm_cognitive_deployment" "eastus2_provisioned" {
  for_each = {
    "gpt4o_regional" = {
      model    = "gpt-4o"
      version  = "2024-05-13"
      capacity = 2
      sku_name = "ProvisionedManaged"
    }
    "gpt4_global" = {
      model    = "gpt-4"
      version  = "0613"
      capacity = 3
      sku_name = "GlobalProvisionedManaged"
    }
  }

  name                 = "eastus2-provisioned-${each.key}"
  cognitive_account_id = azurerm_cognitive_account.eastus2.id

  model {
    format  = "OpenAI"
    name    = each.value.model
    version = each.value.version
  }

  sku {
    name     = each.value.sku_name
    capacity = each.value.capacity
  }
}

resource "azurerm_cognitive_deployment" "swedencentral_provisioned" {
  for_each = {
    "gpt4o_data_zone" = {
      model    = "gpt-4o"
      version  = "2024-05-13"
      capacity = 1
      sku_name = "DataZoneProvisionedManaged"
    }
    "o3_mini_regional" = {
      model    = "o3-mini"
      version  = "2025-01-31"
      capacity = 2
      sku_name = "ProvisionedManaged"
    }
  }

  name                 = "swedencentral-provisioned-${each.key}"
  cognitive_account_id = azurerm_cognitive_account.swedencentral.id

  model {
    format  = "OpenAI"
    name    = each.value.model
    version = each.value.version
  }

  sku {
    name     = each.value.sku_name
    capacity = each.value.capacity
  }
}

# Comprehensive comparison of all three provisioned SKU types for GPT-4o
resource "azurerm_cognitive_deployment" "provisioned_comparison" {
  for_each = {
    "regional" = {
      sku_name = "ProvisionedManaged"
      capacity = 2
    }
    "global" = {
      sku_name = "GlobalProvisionedManaged"
      capacity = 2
    }
    "data_zone" = {
      sku_name = "DataZoneProvisionedManaged"
      capacity = 2
    }
  }

  name                 = "gpt4o-${each.key}-comparison"
  cognitive_account_id = azurerm_cognitive_account.eastus.id

  model {
    format  = "OpenAI"
    name    = "gpt-4o"
    version = "2024-05-13"
  }

  sku {
    name     = each.value.sku_name
    capacity = each.value.capacity
  }
}
