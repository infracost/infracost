package aws

import (
	"infracost/pkg/resource"

	"github.com/shopspring/decimal"
)

func NewLambdaFunction(address string, region string, rawValues map[string]interface{}) resource.Resource {
	r := resource.NewBaseResource(address, rawValues, true)

	costPerReq := decimal.NewFromFloat(0.2 / 1000000)

	// From attached aws_cloudwatch_event_rule
	// schedule_expression = "rate(1 minute)"
	// Can calculate the rate of requests

	timeoutSec := decimal.NewFromFloat(rawValues["timeout"].(float64))
	maxRequestsPerHour := decimal.NewFromInt(60 * 60).Div(timeoutSec)

	hoursProductFilter := &resource.ProductFilter{
		VendorName:    strPtr("aws"),
		Region:        strPtr(region),
		Service:       strPtr("AWS Lambda"),
		ProductFamily: strPtr("Lambda"),
	}
	priceRange := resource.NewBasePriceRangeComponent("Requests", r, "request", "hour", hoursProductFilter, nil)

	// Completely unbounded example
	// The function can never be triggered or the function can run constantly with a the given timeout (3 sec)
	priceRange.SetPriceRange(&resource.PriceRange{
		Min: decimal.NewFromInt(0).Mul(costPerReq),
		Max: maxRequestsPerHour.Mul(costPerReq),
	})
	r.AddPriceRangeComponent(priceRange)

	return r
}
