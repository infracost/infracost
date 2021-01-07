package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetCloudwatchLogGroupItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_cloudwatch_log_group",
		RFunc: NewCloudwatchLogGroup,
	}
}

func NewCloudwatchLogGroup(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	var gbDataIngestion *decimal.Decimal
	var gbDataStorage *decimal.Decimal
	var gbDataScanned *decimal.Decimal

	if u != nil && u.Get("monthly_gb_data_ingestion").Exists() {
		gbDataIngestion = decimalPtr(decimal.NewFromFloat(u.Get("monthly_gb_data_ingestion").Float()))
	}

	if u != nil && u.Get("monthly_gb_data_storage").Exists() {
		gbDataStorage = decimalPtr(decimal.NewFromFloat(u.Get("monthly_gb_data_storage").Float()))
	}

	if u != nil && u.Get("monthly_gb_data_scanned").Exists() {
		gbDataScanned = decimalPtr(decimal.NewFromFloat(u.Get("monthly_gb_data_scanned").Float()))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Data ingestion",
				Unit:            "GB",
				UnitMultiplier:  1,
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
				UnitMultiplier:  1,
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
				UnitMultiplier:  1,
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
		},
	}
}
