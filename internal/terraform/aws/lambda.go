package aws

import (
	"infracost/pkg/resource"

	"github.com/shopspring/decimal"
)

func placeHolderQuantity(resource resource.Resource) decimal.Decimal {
	quantity := decimal.NewFromInt(1).Div(decimal.NewFromInt(730))
	return quantity
}

func NewLambdaFunction(address string, region string, rawValues map[string]interface{}) resource.Resource {
	r := resource.NewBaseResource(address, rawValues, true)

	hoursProductFilter := &resource.ProductFilter{
		VendorName:    strPtr("aws"),
		Region:        strPtr(region),
		Service:       strPtr("AWS Lambda"),
		ProductFamily: strPtr("Lambda"),
	}
	requestsPlaceHolder := resource.NewBasePriceComponent("Requests", r, "request", "hour", hoursProductFilter, nil)
	requestsPlaceHolder.SetPriceOverrideLabel("coming soon")
	requestsPlaceHolder.SetQuantityMultiplierFunc(placeHolderQuantity)
	r.AddPriceComponent(requestsPlaceHolder)
	return r
}
