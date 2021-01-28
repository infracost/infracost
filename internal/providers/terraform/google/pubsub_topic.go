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
	var messageDeliveryTB, storageGB, snapshotStorageGB *decimal.Decimal

	var monthlyMessages, messageSizeKB, subscriptions, messageRetentionDays, snapshotRetentionDays, monthlySnapshots *int64
	if u != nil {
		if u.Get("monthly_messages").Exists() {
			monthlyMessages = int64Ptr(u.Get("monthly_messages").Int())
		}
		if u.Get("message_size_kb").Exists() {
			messageSizeKB = int64Ptr(u.Get("message_size_kb").Int())
		}
		if u.Get("subscriptions").Exists() {
			subscriptions = int64Ptr(u.Get("subscriptions").Int())
		}
		if u.Get("message_retention_days").Exists() {
			messageRetentionDays = int64Ptr(u.Get("message_retention_days").Int())
		}
		if u.Get("snapshot_retention_days").Exists() {
			snapshotRetentionDays = int64Ptr(u.Get("snapshot_retention_days").Int())
		}
		if u.Get("monthly_snapshots").Exists() {
			monthlySnapshots = int64Ptr(u.Get("monthly_snapshots").Int())
		}
	}

	if monthlyMessages != nil && messageSizeKB != nil && subscriptions != nil {
		// Add 1 to subscriptions for the publishing, convert to TiB, subtract 0.01 as the first 10GB is free
		messageDeliveryTB = decimalPtr(decimal.NewFromInt(*monthlyMessages * *messageSizeKB * (*subscriptions + 1)).Div(decimal.NewFromInt(1024 * 1024 * 1024)).Sub(decimal.NewFromFloat(0.01)))
	}

	if monthlyMessages != nil && messageSizeKB != nil && subscriptions != nil && messageRetentionDays != nil {
		storageGB = decimalPtr(decimal.NewFromInt(*monthlyMessages * *messageSizeKB * *subscriptions * *messageRetentionDays).Div(decimal.NewFromInt(30 * 1024 * 1024)))
	}

	if monthlyMessages != nil && messageSizeKB != nil && subscriptions != nil && snapshotRetentionDays != nil && monthlySnapshots != nil {
		// Needs work
		snapshotStorageGB = decimalPtr(decimal.NewFromFloat(
			0.5 * float64(*monthlySnapshots) * float64(*snapshotRetentionDays) * float64(*monthlyMessages/30) * float64(*messageSizeKB)).Div(decimal.NewFromInt(1024 * 1024)))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Message delivery (publish)",
				Unit:            "TiB",
				UnitMultiplier:  1,
				MonthlyQuantity: messageDeliveryTB,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr("global"),
					Service:       strPtr("Cloud Pub/Sub"),
					ProductFamily: strPtr("ApplicationServices"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "description", Value: strPtr("Message Delivery Basic")}, // cd1eba29e63e40fed467ba3eb7c5a6e6-a17de3a1a22680d865c82275b34c9d21
					},
				},
				PriceFilter: &schema.PriceFilter{
					EndUsageAmount: strPtr(""),
				},
			},
			{
				Name:            "Storage",
				Unit:            "GiB-months",
				UnitMultiplier:  1,
				MonthlyQuantity: storageGB,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr("global"),
					Service:       strPtr("Cloud Pub/Sub"),
					ProductFamily: strPtr("ApplicationServices"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "description", Value: strPtr("Subscriptions retained acknowledged messages")}, // 2f872d150153ed13eee94c2bbcc2b69f-57bc5d148491a8381abaccb21ca6b4e9
					},
				},
				PriceFilter: &schema.PriceFilter{
					EndUsageAmount: strPtr(""),
				},
			},
			{
				Name:            "Snapshot storage",
				Unit:            "GiB-months",
				UnitMultiplier:  1,
				MonthlyQuantity: snapshotStorageGB,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr("global"),
					Service:       strPtr("Cloud Pub/Sub"),
					ProductFamily: strPtr("ApplicationServices"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "description", Value: strPtr("Snapshots message backlog")}, // c70f49a01dcd455baf39b999e45961d3-57bc5d148491a8381abaccb21ca6b4e9
					},
				},
				PriceFilter: &schema.PriceFilter{
					EndUsageAmount: strPtr(""),
				},
			},
		},
	}
}
