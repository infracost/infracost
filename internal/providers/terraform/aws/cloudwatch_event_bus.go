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
	if u != nil && u.Get("monthly_partner_events").Exists() {
		monthlyPartnerEvents = decimalPtr(decimal.NewFromInt(u.Get("monthly_partner_events").Int()))
	}
	var monthlyIngestedEvents *decimal.Decimal
	if u != nil && u.Get("monthly_events_ingested_for_discovery").Exists() {
		monthlyIngestedEvents = decimalPtr(decimal.NewFromInt(u.Get("monthly_events_ingested_for_discovery").Int()))
	}
	var monthlyArchivedEvents *decimal.Decimal
	if u != nil && u.Get("monthly_archived_events_gb").Exists() {
		monthlyArchivedEvents = decimalPtr(decimal.NewFromInt(u.Get("monthly_archived_events_gb").Int()))
	}
	var monthlyArchiveProcessing *decimal.Decimal
	if u != nil && u.Get("monthly_archive_processing_gb").Exists() {
		monthlyArchiveProcessing = decimalPtr(decimal.NewFromInt(u.Get("monthly_archive_processing_gb").Int()))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Custom published events",
				Unit:            "events",
				UnitMultiplier:  1000000,
				MonthlyQuantity: monthlyCustomEvents,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AWSEvents"),
					ProductFamily: strPtr("EventBridge"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "eventType", Value: strPtr("Custom Event")},
					},
				},
			},
			{
				Name:            "Partner published events",
				Unit:            "events",
				UnitMultiplier:  1000000,
				MonthlyQuantity: monthlyPartnerEvents,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AWSEvents"),
					ProductFamily: strPtr("EventBridge"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "eventType", Value: strPtr("Partner Event")},
					},
				},
			},
			{
				Name:            "Events ingested for schema discovery",
				Unit:            "events",
				UnitMultiplier:  1000000,
				MonthlyQuantity: monthlyIngestedEvents,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AWSEvents"),
					ProductFamily: strPtr("EventBridge"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "eventType", Value: strPtr("Discovery Event")},
					},
				},
			},
			{
				Name:            "Archived events",
				Unit:            "GB",
				UnitMultiplier:  1,
				MonthlyQuantity: monthlyArchivedEvents,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AWSEvents"),
					ProductFamily: strPtr("CloudWatch Events"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "eventType", Value: strPtr("Event Storage")},
					},
				},
			},
			{
				Name:            "Archive processing",
				Unit:            "GB",
				UnitMultiplier:  1,
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
		},
	}
}
