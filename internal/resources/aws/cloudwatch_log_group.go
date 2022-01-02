package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type CloudwatchLogGroup struct {
	Address               *string
	Region                *string
	MonthlyDataIngestedGb *float64 `infracost_usage:"monthly_data_ingested_gb"`
	StorageGb             *float64 `infracost_usage:"storage_gb"`
	MonthlyDataScannedGb  *float64 `infracost_usage:"monthly_data_scanned_gb"`
}

var CloudwatchLogGroupUsageSchema = []*schema.UsageItem{{Key: "monthly_data_ingested_gb", ValueType: schema.Float64, DefaultValue: 0}, {Key: "storage_gb", ValueType: schema.Float64, DefaultValue: 0}, {Key: "monthly_data_scanned_gb", ValueType: schema.Float64, DefaultValue: 0}}

func (r *CloudwatchLogGroup) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *CloudwatchLogGroup) BuildResource() *schema.Resource {
	region := *r.Region

	var gbDataIngestion *decimal.Decimal
	var gbDataStorage *decimal.Decimal
	var gbDataScanned *decimal.Decimal

	if r.MonthlyDataIngestedGb != nil {
		gbDataIngestion = decimalPtr(decimal.NewFromFloat(*r.MonthlyDataIngestedGb))
	}

	if r.StorageGb != nil {
		gbDataStorage = decimalPtr(decimal.NewFromFloat(*r.StorageGb))
	}

	if r.MonthlyDataScannedGb != nil {
		gbDataScanned = decimalPtr(decimal.NewFromFloat(*r.MonthlyDataScannedGb))
	}

	return &schema.Resource{
		Name: *r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Data ingested",
				Unit:            "GB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: gbDataIngestion,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonCloudWatch"),
					ProductFamily: strPtr("Data Payload"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/-DataProcessing-Bytes/")},
					},
				},
			},
			{
				Name:            "Archival Storage",
				Unit:            "GB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: gbDataStorage,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonCloudWatch"),
					ProductFamily: strPtr("Storage Snapshot"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/-TimedStorage-ByteHrs/")},
					},
				},
			},
			{
				Name:            "Insights queries data scanned",
				Unit:            "GB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: gbDataScanned,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonCloudWatch"),
					ProductFamily: strPtr("Data Payload"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/-DataScanned-Bytes/")},
					},
				},
			},
		}, UsageSchema: CloudwatchLogGroupUsageSchema,
	}
}
