package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type PubSubTopic struct {
	Address              string
	MonthlyMessageDataTB *float64 `infracost_usage:"monthly_message_data_tb"`
}

func (r *PubSubTopic) CoreType() string {
	return "PubSubTopic"
}

func (r *PubSubTopic) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_message_data_tb", ValueType: schema.Float64, DefaultValue: 0.0},
	}
}

func (r *PubSubTopic) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *PubSubTopic) BuildResource() *schema.Resource {
	var messageDataTB *decimal.Decimal

	if r.MonthlyMessageDataTB != nil {
		messageDataTB = decimalPtr(decimal.NewFromFloat(*r.MonthlyMessageDataTB))
	}

	return &schema.Resource{
		Name: r.Address,
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
				UsageBased: true,
			},
		},
		UsageSchema: r.UsageSchema(),
	}
}
