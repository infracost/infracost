package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"

	"github.com/shopspring/decimal"
)

type SNSTopicSubscription struct {
	Address         string
	Protocol        string
	Region          string
	RequestSizeKB   *float64 `infracost_usage:"request_size_kb"`
	MonthlyRequests *int64   `infracost_usage:"monthly_requests"`
}

func (r *SNSTopicSubscription) CoreType() string {
	return "SNSTopicSubscription"
}

func (r *SNSTopicSubscription) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "request_size_kb", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "monthly_requests", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *SNSTopicSubscription) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *SNSTopicSubscription) BuildResource() *schema.Resource {
	var endpointType string
	var freeTier string
	switch r.Protocol {
	case "http", "https":
		endpointType = "HTTP"
		freeTier = "100000"
	default:
		return &schema.Resource{
			Name:      r.Address,
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	var requests *decimal.Decimal

	requestSize := decimal.NewFromInt(64)
	if r.RequestSizeKB != nil {
		requestSize = decimal.NewFromFloat(*r.RequestSizeKB)
	}

	if r.MonthlyRequests != nil {
		requests = decimalPtr(r.calculateRequests(requestSize, decimal.NewFromInt(*r.MonthlyRequests)))
	}

	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            fmt.Sprintf("%s notifications", endpointType),
				Unit:            "1M notifications",
				UnitMultiplier:  decimal.NewFromInt(1000000),
				MonthlyQuantity: requests,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(r.Region),
					Service:       strPtr("AmazonSNS"),
					ProductFamily: strPtr("Message Delivery"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "endpointType", ValueRegex: strPtr(fmt.Sprintf("/%s/i", endpointType))},
					},
				},
				PriceFilter: &schema.PriceFilter{
					StartUsageAmount: strPtr(freeTier),
				},
				UsageBased: true,
			},
		},
		UsageSchema: r.UsageSchema(),
	}
}

func (r *SNSTopicSubscription) calculateRequests(requestSize decimal.Decimal, monthlyRequests decimal.Decimal) decimal.Decimal {
	return requestSize.Div(decimal.NewFromInt(64)).Ceil().Mul(monthlyRequests)
}
