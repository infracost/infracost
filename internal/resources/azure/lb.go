package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"strings"

	"github.com/shopspring/decimal"
)

type LB struct {
	Address                string
	Region                 string
	SKU                    string
	MonthlyDataProcessedGB *float64 `infracost_usage:"monthly_data_processed_gb"`
}

func (r *LB) CoreType() string {
	return "LB"
}

func (r *LB) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{{Key: "monthly_data_processed_gb", ValueType: schema.Float64, DefaultValue: 0}}
}

func (r *LB) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *LB) BuildResource() *schema.Resource {
	region := convertRegion(r.Region)

	var costComponents []*schema.CostComponent
	sku := "Basic"
	var monthlyDataProcessedGb *decimal.Decimal

	if r.MonthlyDataProcessedGB != nil {
		monthlyDataProcessedGb = decimalPtr(decimal.NewFromFloat(*r.MonthlyDataProcessedGB))
	}

	if r.SKU != "" {
		sku = r.SKU
	}

	if strings.ToLower(sku) == "basic" {
		return &schema.Resource{
			Name:        r.Address,
			NoPrice:     true,
			IsSkipped:   true,
			UsageSchema: r.UsageSchema(),
		}
	}

	costComponents = append(costComponents, lbDataProcessedCostComponent(region, monthlyDataProcessedGb))

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func lbDataProcessedCostComponent(region string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Data processed",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Load Balancer"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr("Standard")},
				{Key: "meterName", ValueRegex: regexPtr("Data Processed$")},
			},
		},
		UsageBased: true,
	}
}
