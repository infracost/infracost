package google

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetCloudFunctionsRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_cloudfunctions_function",
		RFunc: NewCloudFunctions,
	}
}

func NewCloudFunctions(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	var monthlyGBzSeconds *decimal.Decimal
	if u != nil && u.Get("monthly_gbz_secs").Exists() {
		monthlyGBzSeconds = decimalPtr(decimal.NewFromInt(u.Get("monthly_gbz_secs").Int()))
	}

	var monthlyGBSeconds *decimal.Decimal
	if u != nil && u.Get("monthly_gb_secs").Exists() {
		monthlyGBSeconds = decimalPtr(decimal.NewFromInt(u.Get("monthly_gb_secs").Int()))
	}

	var invocations *decimal.Decimal
	if u != nil && u.Get("monthly_functions_calls").Exists() {
		invocations = decimalPtr(decimal.NewFromInt(u.Get("monthly_functions_calls").Int()))
	}

	var networkEgrees *decimal.Decimal
	if u != nil && u.Get("monthly_outbound_data_gb").Exists() {
		networkEgrees = decimalPtr(decimal.NewFromInt(u.Get("monthly_outbound_data_gb").Int()))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "CPU Time",
				Unit:            "GBz-Second",
				UnitMultiplier:  1,
				MonthlyQuantity: monthlyGBzSeconds,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr(region),
					Service:       strPtr("Cloud Functions"),
					ProductFamily: strPtr("ApplicationServices"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "description", Value: strPtr("CPU Time")},
					},
				},
			},
			{
				Name:            "Memory Time",
				Unit:            "GB-Second",
				UnitMultiplier:  1,
				MonthlyQuantity: monthlyGBSeconds,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr(region),
					Service:       strPtr("Cloud Functions"),
					ProductFamily: strPtr("ApplicationServices"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "description", Value: strPtr("Memory Time")},
					},
				},
			},
			{
				Name:            "Invocations",
				Unit:            "calls",
				UnitMultiplier:  1,
				MonthlyQuantity: invocations,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr("global"),
					Service:       strPtr("Cloud Functions"),
					ProductFamily: strPtr("ApplicationServices"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "description", Value: strPtr("Invocations")},
					},
				},
				PriceFilter: &schema.PriceFilter{
					StartUsageAmount: strPtr("2000000"), // use the non-free tier
				},
			},
			{
				Name:            "Outbound Data",
				Unit:            "GB",
				UnitMultiplier:  1,
				MonthlyQuantity: networkEgrees,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr("us-east1"),
					Service:       strPtr("Cloud Functions"),
					ProductFamily: strPtr("ApplicationServices"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "description", ValueRegex: strPtr("/Network Egress/")},
					},
				},
				PriceFilter: &schema.PriceFilter{
					StartUsageAmount: strPtr("5"), // use the non-free tier
				},
			},
		},
	}
}
