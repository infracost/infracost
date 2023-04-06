provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "eastus"
}

data "azurerm_client_config" "current" {
}

resource "azurerm_monitor_action_group" "everything_example" {
  for_each = toset(["with_usage", "without_usage"])

  name                = "NotifyEverything"
  resource_group_name = azurerm_resource_group.example.name
  short_name          = "ne"

  arm_role_receiver {
    name                    = "armroleaction"
    role_id                 = "de139f84-1756-47ae-9be6-808fbbe84772"
    use_common_alert_schema = true
  }

  automation_runbook_receiver {
    name                    = "action_name_1"
    automation_account_id   = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-runbooks/providers/Microsoft.Automation/automationAccounts/aaa001"
    runbook_name            = "my runbook"
    webhook_resource_id     = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-runbooks/providers/Microsoft.Automation/automationAccounts/aaa001/webhooks/webhook_alert"
    is_global_runbook       = true
    service_uri             = "https://s13events.azure-automation.net/webhooks?token=randomtoken"
    use_common_alert_schema = true
  }

  azure_app_push_receiver {
    name          = "pushtoadmin"
    email_address = "admin@contoso.com"
  }

  azure_function_receiver {
    name                     = "funcaction"
    function_app_resource_id = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-funcapp/providers/Microsoft.Web/sites/funcapp"
    function_name            = "myfunc"
    http_trigger_url         = "https://example.com/trigger"
    use_common_alert_schema  = true
  }

  email_receiver {
    name          = "sendtoadmin"
    email_address = "admin@contoso.com"
  }

  email_receiver {
    name                    = "sendtodevops"
    email_address           = "devops@contoso.com"
    use_common_alert_schema = true
  }

  itsm_receiver {
    name                 = "createorupdateticket"
    workspace_id         = "00000000-0000-0000-0000-000000000000|00000000-0000-0000-0000-000000000000"
    connection_id        = "53de6956-42b4-41ba-be3c-b154cdf17b13"
    ticket_configuration = "{\"PayloadRevision\":0,\"WorkItemType\":\"Incident\",\"UseTemplate\":false,\"WorkItemData\":\"{}\",\"CreateOneWIPerCI\":false}"
    region               = "southcentralus"
  }

  logic_app_receiver {
    name                    = "logicappaction"
    resource_id             = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-logicapp/providers/Microsoft.Logic/workflows/logicapp"
    callback_url            = "https://logicapptriggerurl/..."
    use_common_alert_schema = true
  }

  sms_receiver {
    name         = "oncallmsg"
    country_code = "1"
    phone_number = "1231231234"
  }

  voice_receiver {
    name         = "remotesupport"
    country_code = "86"
    phone_number = "13888888888"
  }

  webhook_receiver {
    name                    = "callmyapiaswell1"
    service_uri             = "http://example.com/alert"
    use_common_alert_schema = true
  }

  webhook_receiver {
    name                    = "callmyapiaswell2"
    service_uri             = "http://example.com/alert"
    use_common_alert_schema = true
  }

  webhook_receiver {
    name                    = "callmyapiaswell3"
    service_uri             = "http://example.com/alert"
    use_common_alert_schema = true
    aad_auth {
      object_id = "00000000-0000-0000-0000-000000000000"
    }
  }
}

resource "azurerm_monitor_action_group" "partial_example" {
  for_each = toset(["with_usage", "without_usage"])

  name                = "NotifySomethings"
  resource_group_name = azurerm_resource_group.example.name
  short_name          = "ps"

  sms_receiver {
    name         = "oncallmsg"
    country_code = "1"
    phone_number = "1231231234"
  }

  sms_receiver {
    name         = "oncallmsg"
    country_code = "34"
    phone_number = "1231231234"
  }

  voice_receiver {
    name         = "remotesupport"
    country_code = "86"
    phone_number = "13888888888"
  }

  voice_receiver {
    name         = "remotesupport"
    country_code = "64"
    phone_number = "13888888888"
  }

  webhook_receiver {
    name                    = "callmyapiaswell"
    service_uri             = "http://example.com/alert"
    use_common_alert_schema = true
  }
}
