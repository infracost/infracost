package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getMonitorDiagnosticSettingRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_monitor_diagnostic_setting",
		CoreRFunc: newMonitorDiagnosticSetting,
		ReferenceAttributes: []string{
			"target_resource_id",
		},
	}
}

func newMonitorDiagnosticSetting(d *schema.ResourceData) schema.CoreResource {
	region := lookupRegion(d, []string{"target_resource_id"})
	return &azure.MonitorDiagnosticSetting{
		Address: d.Address,
		Region:  region,

		EventHubTarget:        !d.IsEmpty("eventhub_authorization_rule_id"),
		PartnerSolutionTarget: !d.IsEmpty("partner_solution_id"),
		StorageAccountTarget:  !d.IsEmpty("storage_account_id"),
	}
}
