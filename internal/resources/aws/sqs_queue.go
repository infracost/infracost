package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type SQSQueue struct {
	Address         string
	Region          string
	FifoQueue       bool
	MonthlyRequests *float64 `infracost_usage:"monthly_requests"`
	RequestSizeKB   *int64   `infracost_usage:"request_size_kb"`
}

func (r *SQSQueue) CoreType() string {
	return "SQSQueue"
}

func (r *SQSQueue) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_requests", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "request_size_kb", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *SQSQueue) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *SQSQueue) BuildResource() *schema.Resource {
	var queueType string
	if r.FifoQueue {
		queueType = "FIFO (first-in, first-out)"
	} else {
		queueType = "Standard"
	}

	var requests *decimal.Decimal

	requestSize := decimal.NewFromInt(64)
	if r.RequestSizeKB != nil {
		requestSize = decimal.NewFromInt(*r.RequestSizeKB)
	}

	if r.MonthlyRequests != nil {
		requests = decimalPtr(r.calculateRequests(requestSize, decimal.NewFromFloat(*r.MonthlyRequests)))
	}

	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Requests",
				Unit:            "1M requests",
				UnitMultiplier:  decimal.NewFromInt(1000000),
				MonthlyQuantity: requests,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(r.Region),
					Service:       strPtr("AWSQueueService"),
					ProductFamily: strPtr("API Request"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "queueType", Value: strPtr(queueType)},
					},
				},
				PriceFilter: &schema.PriceFilter{
					StartUsageAmount: strPtr("0"),
				},
				UsageBased: true,
			},
		},
		UsageSchema: r.UsageSchema(),
	}
}

func (r *SQSQueue) calculateRequests(requestSize decimal.Decimal, monthlyRequests decimal.Decimal) decimal.Decimal {
	return requestSize.Div(decimal.NewFromInt(64)).Ceil().Mul(monthlyRequests)
}
