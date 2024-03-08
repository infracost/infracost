package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type ComputeTargetGRPCProxy struct {
	Address                string
	Region                 string
	MonthlyProxyInstances  *float64 `infracost_usage:"monthly_proxy_instances"`
	MonthlyDataProcessedGB *float64 `infracost_usage:"monthly_data_processed_gb"`
}

func (r *ComputeTargetGRPCProxy) CoreType() string {
	return "ComputeTargetGRPCProxy"
}

func (r *ComputeTargetGRPCProxy) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{{Key: "monthly_proxy_instances", ValueType: schema.Float64, DefaultValue: 0.000000}, {Key: "monthly_data_processed_gb", ValueType: schema.Float64, DefaultValue: 0}}
}

func (r *ComputeTargetGRPCProxy) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ComputeTargetGRPCProxy) BuildResource() *schema.Resource {
	var monthlyProxyInstances, monthlyDataProcessedGb *decimal.Decimal
	region := r.Region
	costComponents := make([]*schema.CostComponent, 0)

	if r.MonthlyProxyInstances != nil {
		monthlyProxyInstances = decimalPtr(decimal.NewFromFloat(*r.MonthlyProxyInstances))
	}

	costComponents = append(costComponents, r.proxyInstanceCostComponent(monthlyProxyInstances))

	if r.MonthlyDataProcessedGB != nil {
		monthlyDataProcessedGb = decimalPtr(decimal.NewFromFloat(*r.MonthlyDataProcessedGB))
	}

	costComponents = append(costComponents, dataProcessedCostComponent(region, monthlyDataProcessedGb))

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *ComputeTargetGRPCProxy) proxyInstanceCostComponent(instanceCount *decimal.Decimal) *schema.CostComponent {
	var quantity *decimal.Decimal
	if instanceCount != nil {
		instanceHours := schema.HourToMonthUnitMultiplier.Mul(*instanceCount)
		quantity = &instanceHours
	}

	return &schema.CostComponent{
		Name:            "Proxy instance",
		Unit:            "hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Network"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: strPtr("/^Network Load Balancing: Forwarding Rule Minimum/i")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("OnDemand"),
		},
		UsageBased: true,
	}
}

func dataProcessedCostComponent(region string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Data processed",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(region),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Network"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: strPtr("/^Network Internal Load Balancing: Data Processing/i")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("OnDemand"),
		},
		UsageBased: true,
	}
}
