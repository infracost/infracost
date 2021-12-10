package google

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetComputeForwardingRuleRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_compute_forwarding_rule",
		RFunc: NewComputeForwarding,
		Notes: []string{"Price for additional forwarding rule is used"},
	}
}

func GetComputeGlobalForwardingRuleRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_compute_global_forwarding_rule",
		RFunc: NewComputeForwarding,
		Notes: []string{"Price for additional forwarding rule is used"},
	}
}

func NewComputeForwarding(ctx *config.ProjectContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var monthlyIngressDataGb *decimal.Decimal
	region := d.Get("region").String()
	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, computeForwardingCostComponent(region))

	if u != nil && u.Get("monthly_ingress_data_gb").Type != gjson.Null {
		monthlyIngressDataGb = decimalPtr(decimal.NewFromInt(u.Get("monthly_ingress_data_gb").Int()))
	}

	costComponents = append(costComponents, computeIngressDataCostComponent(region, monthlyIngressDataGb))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func computeForwardingCostComponent(region string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Forwarding rules",
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(region),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Network"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: strPtr("/^Network Load Balancing: Forwarding Rule Additional/i")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("OnDemand"),
		},
	}
}

func computeIngressDataCostComponent(region string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Ingress data",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(region),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Network"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: strPtr("/^Network Load Balancing: Data Processing Charge/i")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("OnDemand"),
		},
	}
}
