package aws

import (
	"fmt"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetSNSTopicSubscriptionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_sns_topic_subscription",
		RFunc: NewSnsTopicSubscription,
		Notes: []string{
			"SMS and mobile push not yet supported.",
		},
	}
}

func NewSnsTopicSubscription(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	var endpointType string
	var freeTier string

	requestSize := decimal.NewFromInt(64)
	if u != nil && u.Get("request_size.0.value").Exists() {
		requestSize = decimal.NewFromFloat(u.Get("request_size.0.value").Float())
	}

	var requests *decimal.Decimal
	if u != nil && u.Get("monthly_requests.0.value").Exists() {
		monthlyRequests := decimal.NewFromInt(u.Get("monthly_requests.0.value").Int())
		requests = decimalPtr(calculateRequests(requestSize, monthlyRequests))
	}

	switch d.Get("protocol").String() {
	case "http", "https":
		endpointType = "HTTP"
		freeTier = "100000"
	default:
		return nil
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            fmt.Sprintf("%s notifications", endpointType),
				Unit:            "notifications",
				UnitMultiplier:  1000000,
				MonthlyQuantity: requests,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonSNS"),
					ProductFamily: strPtr("Message Delivery"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "endpointType", Value: strPtr(endpointType)},
					},
				},
				PriceFilter: &schema.PriceFilter{
					StartUsageAmount: strPtr(freeTier),
				},
			},
		},
	}
}
