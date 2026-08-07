package azure

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

type AutomationWatcher struct {
	Address string
	Region  string
}

func (r *AutomationWatcher) CoreType() string {
	return "AutomationWatcher"
}

func (r *AutomationWatcher) UsageSchema() []*schema.UsageItem {
	return nil
}

func (r *AutomationWatcher) PopulateUsage(u *schema.UsageData) {}

func (r *AutomationWatcher) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		{
			Name:           "Watcher",
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(r.Region),
				Service:       strPtr("Automation"),
				ProductFamily: strPtr("Management and Governance"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "meterName", ValueRegex: strPtr("/^Watcher$/i")},
					{Key: "skuName", ValueRegex: strPtr("/^Basic$/i")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption:   strPtr("Consumption"),
				StartUsageAmount: strPtr("744"),
			},
		},
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
	}
}
