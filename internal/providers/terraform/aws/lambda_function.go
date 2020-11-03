package aws

import (
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetLambdaFunctionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_lambda_function",
		Notes: []string{"Provisioned concurrency is not yet supported."},
		RFunc: NewLambdaFunction,
	}
}

func NewLambdaFunction(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	memorySize := decimal.NewFromInt(128)
	if d.Get("memory_size").Exists() {
		memorySize = decimal.NewFromInt(d.Get("memory_size").Int())
	}

	averageRequestDuration := decimal.NewFromInt(100)
	if u != nil && u.Get("average_request_duration.0.value").Exists() {
		averageRequestDuration = decimal.NewFromFloat(u.Get("average_request_duration.0.value").Float())
	}

	var monthlyRequests *decimal.Decimal
	var gbSeconds *decimal.Decimal

	if u != nil && u.Get("monthly_requests.0.value").Exists() {
		monthlyRequests = decimalPtr(decimal.NewFromFloat(u.Get("monthly_requests.0.value").Float()))
		gbSeconds = decimalPtr(calculateGBSeconds(memorySize, averageRequestDuration, *monthlyRequests))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Requests",
				Unit:            "requests",
				UnitMultiplier:  1000000,
				MonthlyQuantity: monthlyRequests,
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
				UnitMultiplier:  1,
				MonthlyQuantity: gbSeconds,
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

func calculateGBSeconds(memorySize decimal.Decimal, averageRequestDuration decimal.Decimal, monthlyRequests decimal.Decimal) decimal.Decimal {
	gb := memorySize.Div(decimal.NewFromInt(1024))
	seconds := averageRequestDuration.Div(decimal.NewFromInt(100)).Ceil().Mul(decimal.NewFromFloat(0.1)) // Round up to closest 100ms and convert to seconds
	return monthlyRequests.Mul(gb).Mul(seconds)
}
