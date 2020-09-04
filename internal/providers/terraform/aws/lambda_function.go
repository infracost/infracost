package aws

import (
	"infracost/pkg/schema"

	"github.com/shopspring/decimal"
)

func NewLambdaFunction(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	monthlyRequests := decimal.Zero
	if u != nil && u.Get("monthly_requests.0.value").Exists() {
		monthlyRequests = decimal.NewFromFloat(u.Get("monthly_requests.0.value").Float())
	}

	memorySize := decimal.NewFromInt(128)
	if d.Get("memory_size").Exists() {
		memorySize = decimal.NewFromInt(d.Get("memory_size").Int())
	}

	requestDuration := decimal.Zero
	if u != nil && u.Get("request_duration.0.value").Exists() {
		requestDuration = decimal.NewFromFloat(u.Get("request_duration.0.value").Float())
	}

	gbSeconds := calculateGBSeconds(memorySize, requestDuration, monthlyRequests)

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Requests",
				Unit:            "requests",
				MonthlyQuantity: &monthlyRequests,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AWSLambda"),
					ProductFamily: strPtr("Serverless"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "group", Value: strPtr("AWS-Lambda-Requests")},
						{Key: "usagetype", ValueRegex: strPtr("/Request/")},
					},
				},
			},
			{
				Name:            "Duration",
				Unit:            "GB-seconds",
				MonthlyQuantity: &gbSeconds,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AWSLambda"),
					ProductFamily: strPtr("Serverless"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "group", Value: strPtr("AWS-Lambda-Duration")},
						{Key: "usagetype", ValueRegex: strPtr("/GB-Second/")},
					},
				},
			},
		},
	}
}

func calculateGBSeconds(memorySize decimal.Decimal, requestDuration decimal.Decimal, monthlyRequests decimal.Decimal) decimal.Decimal {
	gb := memorySize.Div(decimal.NewFromInt(1024))
	seconds := requestDuration.Div(decimal.NewFromInt(100)).Ceil().Mul(decimal.NewFromFloat(0.1)) // Round up to closest 100ms and convert to seconds
	return monthlyRequests.Mul(gb).Mul(seconds)
}
