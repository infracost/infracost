package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type SNSTopic struct {
	Address         string
	Region          string
	RequestSizeKB   *float64 `infracost_usage:"request_size_kb"`
	MonthlyRequests *int64   `infracost_usage:"monthly_requests"`
}

var SNSTopicUsageSchema = []*schema.UsageItem{
	{Key: "request_size_kb", ValueType: schema.Float64, DefaultValue: 0},
	{Key: "monthly_requests", ValueType: schema.Int64, DefaultValue: 0},
}

// This is an experiment to see if using an explicit structure to define the cost components
// can enable anything interesting (e.g. list what cost components could apply to a resource
// without having any IaAC)
func (r *SNSTopic) CostComponents() []*schema.CostComponent {
	return []*schema.CostComponent{
		r.APIRequestsCostComponent(nil),
	}
}

func (r *SNSTopic) APIRequestsCostComponent(requests *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Requests",
		Unit:            "1M requests",
		UnitMultiplier:  decimal.NewFromInt(1000000),
		MonthlyQuantity: requests,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonSNS"),
			ProductFamily: strPtr("API Request"),
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("1000000"),
		},
	}
}

func (r *SNSTopic) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *SNSTopic) BuildResource() *schema.Resource {
	var requests *decimal.Decimal

	requestSize := decimal.NewFromInt(64)
	if r.RequestSizeKB != nil {
		requestSize = decimal.NewFromFloat(*r.RequestSizeKB)
	}

	if r.MonthlyRequests != nil {
		requests = decimalPtr(r.calculateRequests(requestSize, decimal.NewFromInt(*r.MonthlyRequests)))
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: []*schema.CostComponent{r.APIRequestsCostComponent(requests)},
		UsageSchema:    SNSTopicUsageSchema,
	}
}

func (r *SNSTopic) calculateRequests(requestSize decimal.Decimal, monthlyRequests decimal.Decimal) decimal.Decimal {
	return requestSize.Div(decimal.NewFromInt(64)).Ceil().Mul(monthlyRequests)
}
