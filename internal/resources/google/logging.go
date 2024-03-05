package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type Logging struct {
	Address              string
	MonthlyLoggingDataGB *float64 `infracost_usage:"monthly_logging_data_gb"`
}

func (r *Logging) CoreType() string {
	return "Logging"
}

func (r *Logging) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_logging_data_gb", ValueType: schema.Float64, DefaultValue: 0},
	}
}

func (r *Logging) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *Logging) BuildResource() *schema.Resource {
	var loggingDataGB *decimal.Decimal
	if r.MonthlyLoggingDataGB != nil {
		loggingDataGB = decimalPtr(decimal.NewFromFloat(*r.MonthlyLoggingDataGB))
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: []*schema.CostComponent{r.loggingDataCostComponent(loggingDataGB)},
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *Logging) loggingDataCostComponent(quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Logging data",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr("global"),
			Service:       strPtr("Cloud Logging"),
			ProductFamily: strPtr("ApplicationServices"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", Value: strPtr("Log Volume")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("50"),
		},
		UsageBased: true,
	}
}
