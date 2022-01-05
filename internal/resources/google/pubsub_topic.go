package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type PubsubTopic struct {
	Address              *string
	MonthlyMessageDataTb *float64 `infracost_usage:"monthly_message_data_tb"`
}

var PubsubTopicUsageSchema = []*schema.UsageItem{{Key: "monthly_message_data_tb", ValueType: schema.Float64, DefaultValue: 0.000000}}

func (r *PubsubTopic) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *PubsubTopic) BuildResource() *schema.Resource {
	var messageDataTB *decimal.Decimal

	if r.MonthlyMessageDataTb != nil {
		messageDataTB = decimalPtr(decimal.NewFromFloat(*r.MonthlyMessageDataTb))
	}

	return &schema.Resource{
		Name: *r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Message ingestion data",
				Unit:            "TiB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: messageDataTB,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr("global"),
					Service:       strPtr("Cloud Pub/Sub"),
					ProductFamily: strPtr("ApplicationServices"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "description", Value: strPtr("Message Delivery Basic")},
					},
				},
				PriceFilter: &schema.PriceFilter{
					EndUsageAmount: strPtr(""),
				},
			},
		}, UsageSchema: PubsubTopicUsageSchema,
	}
}
