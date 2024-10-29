package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type CloudwatchEventBus struct {
	Address                      string
	Region                       string
	MonthlySchemaDiscoveryEvents *int64   `infracost_usage:"monthly_schema_discovery_events"`
	MonthlyCustomEvents          *int64   `infracost_usage:"monthly_custom_events"`
	MonthlyThirdPartyEvents      *int64   `infracost_usage:"monthly_third_party_events"`
	MonthlyArchiveProcessingGB   *float64 `infracost_usage:"monthly_archive_processing_gb"`
	ArchiveStorageGB             *float64 `infracost_usage:"archive_storage_gb"`
}

func (r *CloudwatchEventBus) CoreType() string {
	return "CloudwatchEventBus"
}

func (r *CloudwatchEventBus) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_schema_discovery_events", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_custom_events", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_third_party_events", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_archive_processing_gb", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "archive_storage_gb", ValueType: schema.Float64, DefaultValue: 0},
	}
}

func (r *CloudwatchEventBus) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *CloudwatchEventBus) BuildResource() *schema.Resource {
	var monthlyCustomEvents *decimal.Decimal
	if r.MonthlyCustomEvents != nil {
		monthlyCustomEvents = decimalPtr(decimal.NewFromInt(*r.MonthlyCustomEvents))
	}
	var monthlyPartnerEvents *decimal.Decimal
	if r.MonthlyThirdPartyEvents != nil {
		monthlyPartnerEvents = decimalPtr(decimal.NewFromInt(*r.MonthlyThirdPartyEvents))
	}
	var monthlyArchiveProcessing *decimal.Decimal
	if r.MonthlyArchiveProcessingGB != nil {
		monthlyArchiveProcessing = decimalPtr(decimal.NewFromFloat(*r.MonthlyArchiveProcessingGB))
	}
	var monthlyArchivedEvents *decimal.Decimal
	if r.ArchiveStorageGB != nil {
		monthlyArchivedEvents = decimalPtr(decimal.NewFromFloat(*r.ArchiveStorageGB))
	}
	var monthlyIngestedEvents *decimal.Decimal
	if r.MonthlySchemaDiscoveryEvents != nil {
		monthlyIngestedEvents = decimalPtr(decimal.NewFromInt(*r.MonthlySchemaDiscoveryEvents))
	}

	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Custom events published",
				Unit:            "1M events",
				UnitMultiplier:  decimal.NewFromInt(1000000),
				MonthlyQuantity: monthlyCustomEvents,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(r.Region),
					Service:       strPtr("AWSEvents"),
					ProductFamily: strPtr("EventBridge"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "eventType", Value: strPtr("Custom Event")},
						{Key: "usagetype", ValueRegex: strPtr("/Event-64K-Chunks/")},
					},
				},
				UsageBased: true,
			},
			{
				Name:            "Third-party events published",
				Unit:            "1M events",
				UnitMultiplier:  decimal.NewFromInt(1000000),
				MonthlyQuantity: monthlyPartnerEvents,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(r.Region),
					Service:       strPtr("AWSEvents"),
					ProductFamily: strPtr("EventBridge"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "eventType", Value: strPtr("Partner Event")},
						{Key: "usagetype", ValueRegex: strPtr("/Event-64K-Chunks/")},
					},
				},
				UsageBased: true,
			},
			{
				Name:            "Archive processing",
				Unit:            "GB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: monthlyArchiveProcessing,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(r.Region),
					Service:       strPtr("AWSEvents"),
					ProductFamily: strPtr("CloudWatch Events"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/ArchivedEvents-Bytes/")},
					},
				},
				UsageBased: true,
			},
			{
				Name:            "Archive storage",
				Unit:            "GB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: monthlyArchivedEvents,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(r.Region),
					Service:       strPtr("AWSEvents"),
					ProductFamily: strPtr("CloudWatch Events"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/TimedStorage-ByteHrs/")},
					},
				},
				UsageBased: true,
			},
			{
				Name:            "Schema discovery",
				Unit:            "1M events",
				UnitMultiplier:  decimal.NewFromInt(1000000),
				MonthlyQuantity: monthlyIngestedEvents,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(r.Region),
					Service:       strPtr("AWSEvents"),
					ProductFamily: strPtr("EventBridge"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "eventType", Value: strPtr("Discovery Event")},
						{Key: "usagetype", ValueRegex: strPtr("/Event-8K-Chunks/")},
					},
				},
				UsageBased: true,
			},
		},
		UsageSchema: r.UsageSchema(),
	}
}
