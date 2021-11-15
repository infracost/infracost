provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-rg"
  location = "westus"
}

resource "azurerm_frontdoor_firewall_policy" "firewall_policy_example" {
  name                = "Example"
  resource_group_name = azurerm_resource_group.example.name

  custom_rule {
    name     = "Rule1"
    enabled  = true
    priority = 1
    type     = "MatchRule"
    action   = "Block"

    match_condition {
      match_variable = "RemoteAddr"
      operator       = "IPMatch"
      match_values   = ["192.168.1.0/24", "10.0.0.0/24"]
    }
  }

  custom_rule {
    name     = "Rule2"
    enabled  = false
    priority = 2
    type     = "MatchRule"
    action   = "Block"

    match_condition {
      match_variable = "RemoteAddr"
      operator       = "IPMatch"
      match_values   = ["192.168.1.0/24"]
    }
  }

  managed_rule {
    type    = "DefaultRuleSet"
    version = "1.0"
  }

  managed_rule {
    type    = "RuleSet2"
    version = "1.0"
  }
}

resource "azurerm_frontdoor_firewall_policy" "firewall_policy_with_usage" {
  name                = "ExampleWithUsage"
  resource_group_name = azurerm_resource_group.example.name

  custom_rule {
    name     = "Rule1"
    enabled  = true
    priority = 1
    type     = "MatchRule"
    action   = "Block"

    match_condition {
      match_variable = "RemoteAddr"
      operator       = "IPMatch"
      match_values   = ["192.168.1.0/24", "10.0.0.0/24"]
    }
  }

  custom_rule {
    name     = "Rule2"
    enabled  = false
    priority = 2
    type     = "MatchRule"
    action   = "Block"

    match_condition {
      match_variable = "RemoteAddr"
      operator       = "IPMatch"
      match_values   = ["192.168.1.0/24"]
    }
  }

  managed_rule {
    type    = "DefaultRuleSet"
    version = "1.0"
  }

  managed_rule {
    type    = "RuleSet2"
    version = "1.0"
  }
}
