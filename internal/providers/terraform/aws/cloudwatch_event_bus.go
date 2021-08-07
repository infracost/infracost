package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetCloudwatchEventBusItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_cloudwatch_event_bus",
		RFunc: NewCloudwatchEventBus,
	}
}

func NewCloudwatchEventBus(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	var monthlyCustomEvents *decimal.Decimal
	if u != nil && u.Get("monthly_custom_events").Exists() {
		monthlyCustomEvents = decimalPtr(decimal.NewFromInt(u.Get("monthly_custom_events").Int()))
	}
	var monthlyPartnerEvents *decimal.Decimal
	if u != nil && u.Get("monthly_third_party_events").Exists() {
		monthlyPartnerEvents = decimalPtr(decimal.NewFromInt(u.Get("monthly_third_party_events").Int()))
	}
	var monthlyArchiveProcessing *decimal.Decimal
	if u != nil && u.Get("monthly_archive_processing_gb").Exists() {
		monthlyArchiveProcessing = decimalPtr(decimal.NewFromInt(u.Get("monthly_archive_processing_gb").Int()))
	}
	var monthlyArchivedEvents *decimal.Decimal
	if u != nil && u.Get("archive_storage_gb").Exists() {
		monthlyArchivedEvents = decimalPtr(decimal.NewFromInt(u.Get("archive_storage_gb").Int()))
	}
	var monthlyIngestedEvents *decimal.Decimal
	if u != nil && u.Get("monthly_schema_discovery_events").Exists() {
		monthlyIngestedEvents = decimalPtr(decimal.NewFromInt(u.Get("monthly_schema_discovery_events").Int()))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Custom events published",
				Unit:            "1M events",
				UnitMultiplier:  decimal.NewFromInt(1000000),
				MonthlyQuantity: monthlyCustomEvents,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AWSEvents"),
					ProductFamily: strPtr("EventBridge"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "eventType", Value: strPtr("Custom Event")},
						{Key: "usagetype", ValueRegex: strPtr("/Event-64K-Chunks/")},
					},
				},
			},
			{
				Name:            "Third-party events published",
				Unit:            "1M events",
				UnitMultiplier:  decimal.NewFromInt(1000000),
				MonthlyQuantity: monthlyPartnerEvents,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AWSEvents"),
					ProductFamily: strPtr("EventBridge"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "eventType", Value: strPtr("Partner Event")},
						{Key: "usagetype", ValueRegex: strPtr("/Event-64K-Chunks/")},
					},
				},
			},
			{
				Name:            "Archive processing",
				Unit:            "GB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: monthlyArchiveProcessing,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AWSEvents"),
					ProductFamily: strPtr("CloudWatch Events"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/ArchivedEvents-Bytes/")},
					},
				},
			},
			{
				Name:            "Archive storage",
				Unit:            "GB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: monthlyArchivedEvents,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AWSEvents"),
					ProductFamily: strPtr("CloudWatch Events"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/TimedStorage-ByteHrs/")},
					},
				},
			},
			{
				Name:            "Schema discovery",
				Unit:            "1M events",
				UnitMultiplier:  decimal.NewFromInt(1000000),
				MonthlyQuantity: monthlyIngestedEvents,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AWSEvents"),
					ProductFamily: strPtr("EventBridge"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "eventType", Value: strPtr("Discovery Event")},
						{Key: "usagetype", ValueRegex: strPtr("/Event-8K-Chunks/")},
					},
				},
			},
		},
	}
}
