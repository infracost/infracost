package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"

	"github.com/shopspring/decimal"
)

type SnsTopicSubscription struct {
	Address         *string
	Protocol        *string
	Region          *string
	RequestSizeKb   *float64 `infracost_usage:"request_size_kb"`
	MonthlyRequests *int64   `infracost_usage:"monthly_requests"`
}

var SnsTopicSubscriptionUsageSchema = []*schema.UsageItem{{Key: "request_size_kb", ValueType: schema.Float64, DefaultValue: 0}, {Key: "monthly_requests", ValueType: schema.Int64, DefaultValue: 0}}

func (r *SnsTopicSubscription) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *SnsTopicSubscription) BuildResource() *schema.Resource {
	region := *r.Region

	var endpointType string
	var freeTier string

	requestSize := decimal.NewFromInt(64)
	if r.RequestSizeKb != nil {
		requestSize = decimal.NewFromFloat(*r.RequestSizeKb)
	}

	var requests *decimal.Decimal
	if r.MonthlyRequests != nil {
		monthlyRequests := decimal.NewFromInt(*r.MonthlyRequests)
		requests = decimalPtr(calculateRequests(requestSize, monthlyRequests))
	}

	switch *r.Protocol {
	case "http", "https":
		endpointType = "HTTP"
		freeTier = "100000"
	default:
		return nil
	}

	return &schema.Resource{
		Name: *r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            fmt.Sprintf("%s notifications", endpointType),
				Unit:            "1M notifications",
				UnitMultiplier:  decimal.NewFromInt(1000000),
				MonthlyQuantity: requests,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonSNS"),
					ProductFamily: strPtr("Message Delivery"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "endpointType", ValueRegex: strPtr(fmt.Sprintf("/%s/i", endpointType))},
					},
				},
				PriceFilter: &schema.PriceFilter{
					StartUsageAmount: strPtr(freeTier),
				},
			},
		}, UsageSchema: SnsTopicSubscriptionUsageSchema,
	}
}
