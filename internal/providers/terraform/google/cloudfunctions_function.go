package google

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetCloudFunctionsRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_cloudfunctions_function",
		RFunc: NewCloudFunctions,
	}
}

func NewCloudFunctions(ctx *config.ProjectContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	memorySize := decimal.NewFromInt(256)
	if d.Get("available_memory_mb").Exists() {
		memorySize = decimal.NewFromInt(d.Get("available_memory_mb").Int())
	}

	var cpuMapping = map[int]decimal.Decimal{
		128:  decimal.NewFromInt(200),
		256:  decimal.NewFromInt(400),
		512:  decimal.NewFromInt(800),
		1024: decimal.NewFromInt(1400),
		2048: decimal.NewFromInt(2400),
		4096: decimal.NewFromInt(4800),
	}

	cpuSize := cpuMapping[int(memorySize.IntPart())]

	requestDuration := decimal.NewFromInt(100)
	if u != nil && u.Get("request_duration_ms").Exists() {
		// Round up to nearest 100ms
		requestDuration = decimal.NewFromInt(u.Get("request_duration_ms").Int()).Div(decimal.NewFromInt(100)).Ceil().Mul(decimal.NewFromFloat(100))
	}

	var invocations, monthlyCPUUsage, monthlyMemoryUsage *decimal.Decimal
	if u != nil && u.Get("monthly_function_invocations").Exists() {
		invocations = decimalPtr(decimal.NewFromInt(u.Get("monthly_function_invocations").Int()))
		monthlyCPUUsage = decimalPtr(calculateGHzSeconds(cpuSize, requestDuration, *invocations))
		monthlyMemoryUsage = decimalPtr(calculateGBSeconds(memorySize, requestDuration, *invocations))
	}

	var networkEgrees *decimal.Decimal
	if u != nil && u.Get("monthly_outbound_data_gb").Exists() {
		networkEgrees = decimalPtr(decimal.NewFromInt(u.Get("monthly_outbound_data_gb").Int()))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "CPU",
				Unit:            "GHz-seconds",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: monthlyCPUUsage,
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
				Name:            "Memory",
				Unit:            "GB-seconds",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: monthlyMemoryUsage,
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
				Unit:            "invocations",
				UnitMultiplier:  decimal.NewFromInt(1),
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
				Name:            "Outbound data transfer",
				Unit:            "GB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: networkEgrees,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr(region),
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

func calculateGBSeconds(memorySize decimal.Decimal, averageRequestDuration decimal.Decimal, monthlyRequests decimal.Decimal) decimal.Decimal {
	gb := memorySize.Div(decimal.NewFromInt(1024))
	seconds := averageRequestDuration.Div(decimal.NewFromInt(1000))
	return monthlyRequests.Mul(gb).Mul(seconds)
}

func calculateGHzSeconds(memorySize decimal.Decimal, averageRequestDuration decimal.Decimal, monthlyRequests decimal.Decimal) decimal.Decimal {
	gb := memorySize.Div(decimal.NewFromInt(1000))
	seconds := averageRequestDuration.Div(decimal.NewFromInt(1000))
	return monthlyRequests.Mul(gb).Mul(seconds)
}
