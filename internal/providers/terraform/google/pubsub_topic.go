package google

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetPubSubTopicRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_pubsub_topic",
		RFunc: NewPubSubTopic,
	}
}

func NewPubSubTopic(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var messageDataTB *decimal.Decimal

	if u != nil && u.Get("monthly_message_data_tb").Exists() {
		messageDataTB = decimalPtr(decimal.NewFromFloat(u.Get("monthly_message_data_tb").Float()))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Message ingestion data",
				Unit:            "TiB",
				UnitMultiplier:  1,
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
		},
	}
}
