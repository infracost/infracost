package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type SNSTopic struct {
	Address         *string
	Region          *string
	RequestSizeKb   *float64 `infracost_usage:"request_size_kb"`
	MonthlyRequests *int64   `infracost_usage:"monthly_requests"`
}

var SNSTopicUsageSchema = []*schema.UsageItem{{Key: "request_size_kb", ValueType: schema.Float64, DefaultValue: 0}, {Key: "monthly_requests", ValueType: schema.Int64, DefaultValue: 0}}

func (r *SNSTopic) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *SNSTopic) BuildResource() *schema.Resource {
	region := *r.Region

	requestSize := decimal.NewFromInt(64)
	if r != nil && r.RequestSizeKb != nil {
		requestSize = decimal.NewFromFloat(*r.RequestSizeKb)
	}

	var requests *decimal.Decimal

	if r != nil && r.MonthlyRequests != nil {
		monthlyRequests := decimal.NewFromInt(*r.MonthlyRequests)
		requests = decimalPtr(calculateRequests(requestSize, monthlyRequests))
	}

	return &schema.Resource{
		Name: *r.Address,
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
		}, UsageSchema: SNSTopicUsageSchema,
	}
}
