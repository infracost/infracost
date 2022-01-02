package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type SqsQueue struct {
	Address         *string
	Region          *string
	FifoQueue       *bool
	MonthlyRequests *float64 `infracost_usage:"monthly_requests"`
	RequestSizeKb   *int64   `infracost_usage:"request_size_kb"`
}

var SqsQueueUsageSchema = []*schema.UsageItem{{Key: "monthly_requests", ValueType: schema.Float64, DefaultValue: 0}, {Key: "request_size_kb", ValueType: schema.Int64, DefaultValue: 0}}

func (r *SqsQueue) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *SqsQueue) BuildResource() *schema.Resource {
	region := *r.Region

	var queueType string

	if *r.FifoQueue {
		queueType = "FIFO (first-in, first-out)"
	} else {
		queueType = "Standard"
	}

	requestSize := decimal.NewFromInt(64)
	if r.RequestSizeKb != nil {
		requestSize = decimal.NewFromInt(*r.RequestSizeKb)
	}

	var requests *decimal.Decimal

	if r.MonthlyRequests != nil {
		monthlyRequests := decimal.NewFromFloat(*r.MonthlyRequests)
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
					Service:       strPtr("AWSQueueService"),
					ProductFamily: strPtr("API Request"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "queueType", Value: strPtr(queueType)},
					},
				},
				PriceFilter: &schema.PriceFilter{
					StartUsageAmount: strPtr("0"),
				},
			},
		}, UsageSchema: SqsQueueUsageSchema,
	}
}

func calculateRequests(requestSize decimal.Decimal, monthlyRequests decimal.Decimal) decimal.Decimal {
	return requestSize.Div(decimal.NewFromInt(64)).Ceil().Mul(monthlyRequests)
}
