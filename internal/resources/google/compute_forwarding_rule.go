package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type ComputeForwardingRule struct {
	Address              string
	Region               string
	MonthlyIngressDataGB *float64 `infracost_usage:"monthly_ingress_data_gb"`
}

func (r *ComputeForwardingRule) CoreType() string {
	return "ComputeForwardingRule"
}

func (r *ComputeForwardingRule) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{{Key: "monthly_ingress_data_gb", ValueType: schema.Float64, DefaultValue: 0}}
}

func (r *ComputeForwardingRule) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ComputeForwardingRule) BuildResource() *schema.Resource {
	var monthlyIngressDataGb *decimal.Decimal
	region := r.Region
	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, r.computeForwardingCostComponent())

	if r.MonthlyIngressDataGB != nil {
		monthlyIngressDataGb = decimalPtr(decimal.NewFromFloat(*r.MonthlyIngressDataGB))
	}

	costComponents = append(costComponents, computeIngressDataCostComponent(region, monthlyIngressDataGb))

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *ComputeForwardingRule) computeForwardingCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Forwarding rules",
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(r.Region),
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
		UsageBased: true,
	}
}
