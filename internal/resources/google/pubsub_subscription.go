package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type PubSubSubscription struct {
	Address              string
	MonthlyMessageDataTB *float64 `infracost_usage:"monthly_message_data_tb"`
	StorageGB            *float64 `infracost_usage:"storage_gb"`
	SnapshotStorageGB    *float64 `infracost_usage:"snapshot_storage_gb"`
}

func (r *PubSubSubscription) CoreType() string {
	return "PubSubSubscription"
}

func (r *PubSubSubscription) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_message_data_tb", ValueType: schema.Float64, DefaultValue: 0.0},
		{Key: "storage_gb", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "snapshot_storage_gb", ValueType: schema.Float64, DefaultValue: 0.0},
	}
}

func (r *PubSubSubscription) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *PubSubSubscription) BuildResource() *schema.Resource {
	var messageDataTB, storageGB, snapshotStorageGB *decimal.Decimal

	if r != nil {
		if r.MonthlyMessageDataTB != nil {
			messageDataTB = decimalPtr(decimal.NewFromFloat(*r.MonthlyMessageDataTB))
		}
		if r.StorageGB != nil {
			storageGB = decimalPtr(decimal.NewFromFloat(*r.StorageGB))
		}
		if r.SnapshotStorageGB != nil {
			snapshotStorageGB = decimalPtr(decimal.NewFromFloat(*r.SnapshotStorageGB))
		}
	}

	return &schema.Resource{
		Name: r.Address,
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
				UsageBased: true,
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
				UsageBased: true,
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
				UsageBased: true,
			},
		},
		UsageSchema: r.UsageSchema(),
	}
}
