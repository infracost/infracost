package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type PubsubSubscription struct {
	Address              *string
	MonthlyMessageDataTb *float64 `infracost_usage:"monthly_message_data_tb"`
	StorageGb            *float64 `infracost_usage:"storage_gb"`
	SnapshotStorageGb    *float64 `infracost_usage:"snapshot_storage_gb"`
}

var PubsubSubscriptionUsageSchema = []*schema.UsageItem{{Key: "monthly_message_data_tb", ValueType: schema.Float64, DefaultValue: 0.000000}, {Key: "storage_gb", ValueType: schema.Float64, DefaultValue: 0}, {Key: "snapshot_storage_gb", ValueType: schema.Float64, DefaultValue: 0.000000}}

func (r *PubsubSubscription) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *PubsubSubscription) BuildResource() *schema.Resource {
	var messageDataTB, storageGB, snapshotStorageGB *decimal.Decimal

	if r != nil {
		if r.MonthlyMessageDataTb != nil {
			messageDataTB = decimalPtr(decimal.NewFromFloat(*r.MonthlyMessageDataTb))
		}
		if r.StorageGb != nil {
			storageGB = decimalPtr(decimal.NewFromFloat(*r.StorageGb))
		}
		if r.SnapshotStorageGb != nil {
			snapshotStorageGB = decimalPtr(decimal.NewFromFloat(*r.SnapshotStorageGb))
		}
	}

	return &schema.Resource{
		Name: *r.Address,
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
		}, UsageSchema: PubsubSubscriptionUsageSchema,
	}
}
