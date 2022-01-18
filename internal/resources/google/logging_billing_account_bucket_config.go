package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type LoggingBillingAccountBucketConfig struct {
	Address              *string
	MonthlyLoggingDataGb *float64 `infracost_usage:"monthly_logging_data_gb"`
}

var LoggingBillingAccountBucketConfigUsageSchema = []*schema.UsageItem{{Key: "monthly_logging_data_gb", ValueType: schema.Float64, DefaultValue: 0}}

func (r *LoggingBillingAccountBucketConfig) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *LoggingBillingAccountBucketConfig) BuildResource() *schema.Resource {
	var loggingData *decimal.Decimal
	if r.MonthlyLoggingDataGb != nil {
		loggingData = decimalPtr(decimal.NewFromFloat(*r.MonthlyLoggingDataGb))
	}

	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: loggingCostComponent(loggingData), UsageSchema: LoggingBillingAccountBucketConfigUsageSchema,
	}
}

func loggingCostComponent(loggingData *decimal.Decimal) []*schema.CostComponent {
	return []*schema.CostComponent{
		{
			Name:            "Logging data",
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: loggingData,
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
		},
	}
}
