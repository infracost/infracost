package google

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetLoggingFolderSinkRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_logging_folder_sink",
		RFunc: NewLoggingFolderSink,
	}
}

func NewLoggingFolderSink(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var loggingData *decimal.Decimal
	if u != nil && u.Get("monthly_logging_data_gb").Exists() {
		loggingData = decimalPtr(decimal.NewFromInt(u.Get("monthly_logging_data_gb").Int()))
	}

	costComponents := []*schema.CostComponent{
		{
			Name:            "Logging data",
			Unit:            "GB-months",
			UnitMultiplier:  1,
			MonthlyQuantity: loggingData,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr("global"),
				Service:       strPtr("Stackdriver Logging"),
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

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
