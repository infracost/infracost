package azure

import (
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// MonitorDataCollectionRule struct represents an Azure Monitor Data Collection Rule.
//
// Resource information: https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/monitor_data_collection_rule
// Pricing information: https://azure.microsoft.com/en-in/pricing/details/monitor/
type MonitorDataCollectionRule struct {
	Address string
	Region  string

	MonthlyCustomMetricsSamplesGB *int64 `infracost_usage:"monthly_custom_metrics_samples"`
}

// CoreType returns the name of this resource type
func (r *MonitorDataCollectionRule) CoreType() string {
	return "MonitorDataCollectionRule"
}

// UsageSchema defines a list which represents the usage schema of MonitorDataCollectionRule.
func (r *MonitorDataCollectionRule) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_custom_metrics_samples", ValueType: schema.Int64, DefaultValue: 0},
	}
}

// PopulateUsage parses the u schema.UsageData into the MonitorDataCollectionRule.
// It uses the `infracost_usage` struct tags to populate data into the MonitorDataCollectionRule.
func (r *MonitorDataCollectionRule) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid MonitorDataCollectionRule struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *MonitorDataCollectionRule) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.metricsIngestionCostComponent(r.MonthlyCustomMetricsSamplesGB),
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *MonitorDataCollectionRule) metricsIngestionCostComponent(quantity *int64) *schema.CostComponent {
	var q *decimal.Decimal
	if quantity != nil {
		q = decimalPtr(decimal.NewFromInt(*quantity).Div(decimal.NewFromInt(10000000)))
	}

	return &schema.CostComponent{
		Name:            "Metrics ingestion",
		Unit:            "10M samples",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Azure Monitor"),
			ProductFamily: strPtr("Management and Governance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr("Metrics ingestion")},
				{Key: "meterName", Value: strPtr("Metrics ingestion Metric samples")},
			},
		},
		UsageBased: true,
	}
}
