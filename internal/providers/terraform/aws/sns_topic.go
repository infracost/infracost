package aws

import (
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetSNSTopicRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_sns_topic",
		RFunc: NewSnsTopic,
	}
}

func NewSnsTopic(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	requestSize := decimal.NewFromInt(64)
	if u != nil && u.Get("request_size.0.value").Exists() {
		requestSize = decimal.NewFromFloat(u.Get("request_size.0.value").Float())
	}

	var requests *decimal.Decimal

	if u != nil && u.Get("monthly_requests.0.value").Exists() {
		monthlyRequests := decimal.NewFromInt(u.Get("monthly_requests.0.value").Int())
		requests = decimalPtr(calculateRequests(requestSize, monthlyRequests))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Requests",
				Unit:            "requests",
				UnitMultiplier:  1000000,
				MonthlyQuantity: requests,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonSNS"),
					ProductFamily: strPtr("API Request"),
				},
				PriceFilter: &schema.PriceFilter{
					StartUsageAmount: strPtr("1000000"),
				},
			},
		},
	}
}
