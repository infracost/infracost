package azure

import (
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// MonitorDiagnosticSetting struct represents an Azure Monitor Diagnostics Setting
//
// Resource information: https://learn.microsoft.com/en-us/azure/azure-monitor/essentials/diagnostic-settings
// Pricing information: https://azure.microsoft.com/en-in/pricing/details/monitor/
type MonitorDiagnosticSetting struct {
	Address string
	Region  string

	EventHubTarget        bool
	PartnerSolutionTarget bool
	StorageAccountTarget  bool

	MonthlyPlatformLogGB *int64 `infracost_usage:"monthly_platform_log_gb"`
}

// CoreType returns the name of this resource type
func (r *MonitorDiagnosticSetting) CoreType() string {
	return "MonitorDiagnosticSetting"
}

// UsageSchema defines a list which represents the usage schema of MonitorDiagnosticSetting.
func (r *MonitorDiagnosticSetting) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_platform_log_gb", ValueType: schema.Int64, DefaultValue: 0},
	}
}

// PopulateUsage parses the u schema.UsageData into the MonitorDiagnosticSetting.
// It uses the `infracost_usage` struct tags to populate data into the MonitorDiagnosticSetting.
func (r *MonitorDiagnosticSetting) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid MonitorDiagnosticSetting struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *MonitorDiagnosticSetting) BuildResource() *schema.Resource {
	var costComponents []*schema.CostComponent

	if r.EventHubTarget || r.PartnerSolutionTarget || r.StorageAccountTarget {
		costComponents = []*schema.CostComponent{r.platformLogComponent(r.MonthlyPlatformLogGB)}
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *MonitorDiagnosticSetting) platformLogComponent(q *int64) *schema.CostComponent {
	var quantity *decimal.Decimal
	if q != nil {
		quantity = decimalPtr(decimal.NewFromInt(*q))
	}

	return &schema.CostComponent{
		Name:            "Platform logs processed",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Azure Monitor"),
			ProductFamily: strPtr("Management and Governance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr("Platform Logs")},
			},
		},
		UsageBased: true,
	}
}
