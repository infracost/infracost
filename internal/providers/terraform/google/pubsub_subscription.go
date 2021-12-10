package google

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetPubSubSubscriptionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_pubsub_subscription",
		RFunc: NewPubSubSubscription,
	}
}

func NewPubSubSubscription(ctx *config.ProjectContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var messageDataTB, storageGB, snapshotStorageGB *decimal.Decimal

	if u != nil {
		if u.Get("monthly_message_data_tb").Exists() {
			messageDataTB = decimalPtr(decimal.NewFromFloat(u.Get("monthly_message_data_tb").Float()))
		}
		if u.Get("storage_gb").Exists() {
			storageGB = decimalPtr(decimal.NewFromFloat(u.Get("storage_gb").Float()))
		}
		if u.Get("snapshot_storage_gb").Exists() {
			snapshotStorageGB = decimalPtr(decimal.NewFromFloat(u.Get("snapshot_storage_gb").Float()))
		}
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Message delivery data",
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
			{
				Name:            "Retained acknowledged message storage",
				Unit:            "GiB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: storageGB,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr("global"),
					Service:       strPtr("Cloud Pub/Sub"),
					ProductFamily: strPtr("ApplicationServices"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "description", Value: strPtr("Subscriptions retained acknowledged messages")},
					},
				},
				PriceFilter: &schema.PriceFilter{
					EndUsageAmount: strPtr(""),
				},
			},
			{
				Name:            "Snapshot message backlog storage",
				Unit:            "GiB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: snapshotStorageGB,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr("global"),
					Service:       strPtr("Cloud Pub/Sub"),
					ProductFamily: strPtr("ApplicationServices"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "description", Value: strPtr("Snapshots message backlog")},
					},
				},
				PriceFilter: &schema.PriceFilter{
					EndUsageAmount: strPtr(""),
				},
			},
		},
	}
}
