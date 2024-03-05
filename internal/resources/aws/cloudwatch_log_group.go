package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type CloudwatchLogGroup struct {
	Address               string
	Region                string
	MonthlyDataIngestedGB *float64 `infracost_usage:"monthly_data_ingested_gb"`
	StorageGB             *float64 `infracost_usage:"storage_gb"`
	MonthlyDataScannedGB  *float64 `infracost_usage:"monthly_data_scanned_gb"`
}

var CloudwatchLogGroupUsageSchema = []*schema.UsageItem{
	{Key: "monthly_data_ingested_gb", ValueType: schema.Float64, DefaultValue: 0},
	{Key: "storage_gb", ValueType: schema.Float64, DefaultValue: 0},
	{Key: "monthly_data_scanned_gb", ValueType: schema.Float64, DefaultValue: 0},
}

func (r *CloudwatchLogGroup) CoreType() string {
	return "CloudwatchLogGroup"
}

func (r *CloudwatchLogGroup) UsageSchema() []*schema.UsageItem {
	return CloudwatchLogGroupUsageSchema
}

func (r *CloudwatchLogGroup) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *CloudwatchLogGroup) BuildResource() *schema.Resource {
	var gbDataIngestion *decimal.Decimal
	var gbDataStorage *decimal.Decimal
	var gbDataScanned *decimal.Decimal

	if r.MonthlyDataIngestedGB != nil {
		gbDataIngestion = decimalPtr(decimal.NewFromFloat(*r.MonthlyDataIngestedGB))
	}

	if r.StorageGB != nil {
		gbDataStorage = decimalPtr(decimal.NewFromFloat(*r.StorageGB))
	}

	if r.MonthlyDataScannedGB != nil {
		gbDataScanned = decimalPtr(decimal.NewFromFloat(*r.MonthlyDataScannedGB))
	}

	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Data ingested",
				Unit:            "GB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: gbDataIngestion,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(r.Region),
					Service:       strPtr("AmazonCloudWatch"),
					ProductFamily: strPtr("Data Payload"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/-DataProcessing-Bytes/")},
					},
				},
				UsageBased: true,
			},
			{
				Name:            "Archival Storage",
				Unit:            "GB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: gbDataStorage,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(r.Region),
					Service:       strPtr("AmazonCloudWatch"),
					ProductFamily: strPtr("Storage Snapshot"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/-TimedStorage-ByteHrs/")},
					},
				},
				UsageBased: true,
			},
			{
				Name:            "Insights queries data scanned",
				Unit:            "GB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: gbDataScanned,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(r.Region),
					Service:       strPtr("AmazonCloudWatch"),
					ProductFamily: strPtr("Data Payload"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/-DataScanned-Bytes/")},
					},
				},
				UsageBased: true,
			},
		},
		UsageSchema: r.UsageSchema(),
	}
}
