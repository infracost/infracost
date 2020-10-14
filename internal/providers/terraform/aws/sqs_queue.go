package aws

import (
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetSQSQueueRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_sqs_queue",
		RFunc: NewSqsQueue,
	}
}

func NewSqsQueue(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	var queueType string

	monthlyRequests := decimal.Zero

	if u != nil && u.Get("monthly_requests.0.value").Exists() {
		monthlyRequests = decimal.NewFromFloat(u.Get("monthly_requests.0.value").Float())
	}

	requestSize := decimal.NewFromInt(64)

	if u != nil && u.Get("request_size.0.value").Exists() {
		requestSize = decimal.NewFromInt(u.Get("request_size.0.value").Int())
	}

	if d.Get("fifo_queue").Bool() {
		queueType = "FIFO (first-in, first-out)"
	} else {
		queueType = "Standard"
	}

	requests := calculateSqsRequests(requestSize, monthlyRequests)

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Requests",
				Unit:            "requests",
				MonthlyQuantity: &requests,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AWSQueueService"),
					ProductFamily: strPtr("API Request"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "queueType", Value: strPtr(queueType)},
					},
				},
			},
		},
	}
}

func calculateSqsRequests(requestSize decimal.Decimal, monthlyRequests decimal.Decimal) decimal.Decimal {
	return requestSize.Div(decimal.NewFromInt(64)).Ceil().Mul(monthlyRequests)
}
