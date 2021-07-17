package aws

import (
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type LambdaFunctionArguments struct {
	Address    string `json:"address,omitempty"`
	Region     string `json:"region,omitempty"`
	MemorySize int64  `json:"memorySize,omitempty"`

	RequestDurationMS *float64 `json:"requestDurationMS,omitempty"`
	MonthlyRequests   *float64 `json:"monthlyRequests,omitempty"`
}

func (args *LambdaFunctionArguments) PopulateUsage(u *schema.UsageData) {
	if u != nil {
		args.RequestDurationMS = u.GetFloat("request_duration_ms")
		args.MonthlyRequests = u.GetFloat("monthly_requests")
	}
}

var LambdaFunctionUsageSchema = []*schema.UsageSchemaItem{
	{Key: "request_duration_ms", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "monthly_requests", DefaultValue: 0, ValueType: schema.Float64},
}

func NewLambdaFunction(args *LambdaFunctionArguments) *schema.Resource {
	memorySize := decimal.NewFromInt(args.MemorySize)

	averageRequestDuration := decimal.NewFromInt(1)
	if args.RequestDurationMS != nil {
		averageRequestDuration = decimal.NewFromFloat(*args.RequestDurationMS)
	}

	var monthlyRequests *decimal.Decimal
	var gbSeconds *decimal.Decimal

	if args.MonthlyRequests != nil {
		monthlyRequests = decimalPtr(decimal.NewFromFloat(*args.MonthlyRequests))
		gbSeconds = decimalPtr(calculateGBSeconds(memorySize, averageRequestDuration, *monthlyRequests))
	}

	return &schema.Resource{
		Name: args.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Requests",
				Unit:            "1M requests",
				UnitMultiplier:  1000000,
				MonthlyQuantity: monthlyRequests,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(args.Region),
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
					Region:        strPtr(args.Region),
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
	seconds := averageRequestDuration.Ceil().Div(decimal.NewFromInt(1000)) // Round up to closest 1ms and convert to seconds
	return monthlyRequests.Mul(gb).Mul(seconds)
}
