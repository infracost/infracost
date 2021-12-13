package aws

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetSNSTopicRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_sns_topic",
		RFunc: NewSnsTopic,
	}
}

func NewSnsTopic(ctx *config.RunContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	requestSize := decimal.NewFromInt(64)
	if u != nil && u.Get("request_size_kb").Exists() {
		requestSize = decimal.NewFromFloat(u.Get("request_size_kb").Float())
	}

	var requests *decimal.Decimal

	if u != nil && u.Get("monthly_requests").Exists() {
		monthlyRequests := decimal.NewFromInt(u.Get("monthly_requests").Int())
		requests = decimalPtr(calculateRequests(requestSize, monthlyRequests))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Requests",
				Unit:            "1M requests",
				UnitMultiplier:  decimal.NewFromInt(1000000),
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
