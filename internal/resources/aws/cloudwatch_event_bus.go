package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type CloudwatchEventBus struct {
	Address                      *string
	Region                       *string
	MonthlySchemaDiscoveryEvents *int64   `infracost_usage:"monthly_schema_discovery_events"`
	MonthlyCustomEvents          *int64   `infracost_usage:"monthly_custom_events"`
	MonthlyThirdPartyEvents      *int64   `infracost_usage:"monthly_third_party_events"`
	MonthlyArchiveProcessingGb   *float64 `infracost_usage:"monthly_archive_processing_gb"`
	ArchiveStorageGb             *float64 `infracost_usage:"archive_storage_gb"`
}

var CloudwatchEventBusUsageSchema = []*schema.UsageItem{{Key: "monthly_schema_discovery_events", ValueType: schema.Int64, DefaultValue: 0}, {Key: "monthly_custom_events", ValueType: schema.Int64, DefaultValue: 0}, {Key: "monthly_third_party_events", ValueType: schema.Int64, DefaultValue: 0}, {Key: "monthly_archive_processing_gb", ValueType: schema.Float64, DefaultValue: 0}, {Key: "archive_storage_gb", ValueType: schema.Float64, DefaultValue: 0}}

func (r *CloudwatchEventBus) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *CloudwatchEventBus) BuildResource() *schema.Resource {
	region := *r.Region

	var monthlyCustomEvents *decimal.Decimal
	if r.MonthlyCustomEvents != nil {
		monthlyCustomEvents = decimalPtr(decimal.NewFromInt(*r.MonthlyCustomEvents))
	}
	var monthlyPartnerEvents *decimal.Decimal
	if r.MonthlyThirdPartyEvents != nil {
		monthlyPartnerEvents = decimalPtr(decimal.NewFromInt(*r.MonthlyThirdPartyEvents))
	}
	var monthlyArchiveProcessing *decimal.Decimal
	if r.MonthlyArchiveProcessingGb != nil {
		monthlyArchiveProcessing = decimalPtr(decimal.NewFromFloat(*r.MonthlyArchiveProcessingGb))
	}
	var monthlyArchivedEvents *decimal.Decimal
	if r.ArchiveStorageGb != nil {
		monthlyArchivedEvents = decimalPtr(decimal.NewFromFloat(*r.ArchiveStorageGb))
	}
	var monthlyIngestedEvents *decimal.Decimal
	if r.MonthlySchemaDiscoveryEvents != nil {
		monthlyIngestedEvents = decimalPtr(decimal.NewFromInt(*r.MonthlySchemaDiscoveryEvents))
	}

	return &schema.Resource{
		Name: *r.Address,
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
		}, UsageSchema: CloudwatchEventBusUsageSchema,
	}
}
