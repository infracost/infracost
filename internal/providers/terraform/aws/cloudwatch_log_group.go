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

func NewCloudwatchLogGroup(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	var gbDataAnalzyedQueries int64

	gbDataIngestion := decimal.NewFromInt(0)
	gbDataStorage := decimal.NewFromInt(0)
	gbDataAnalzyed := decimal.NewFromInt(0)

	if u != nil && u.Get("data_ingestion.0.value").Exists() {
		gbDataIngestion = decimal.NewFromInt(u.Get("data_ingestion.0.value").Int())
	}

	if u != nil && u.Get("data_storage.0.value").Exists() {
		gbDataStorage = decimal.NewFromInt(u.Get("data_storage.0.value").Int())
	}

	if u != nil && u.Get("data_analyzed.0.value").Exists() {
		gbDataAnalzyed = decimal.NewFromInt(u.Get("data_analyzed.0.value").Int())
	}

	if u != nil && u.Get("data_queries.0.value").Exists() {
		gbDataAnalzyedQueries = u.Get("data_queries.0.value").Int()
	}

	costComponents := []*schema.CostComponent{
		{
			Name:            "Collect (Data Ingestion)",
			Unit:            "GB",
			UnitMultiplier:  1,
			MonthlyQuantity: decimalPtr(gbDataIngestion),
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
			Name:            "Store (Archival)",
			Unit:            "GB",
			UnitMultiplier:  1,
			MonthlyQuantity: decimalPtr(gbDataStorage),
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
	}

	if gbDataAnalzyedQueries > 0 && gbDataAnalzyed.GreaterThan(decimal.NewFromInt(0)) {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Analyze (Logs insights queries)",
			Unit:            "GB-scanned",
			UnitMultiplier:  int(gbDataAnalzyedQueries),
			MonthlyQuantity: decimalPtr(gbDataAnalzyed),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonCloudWatch"),
				ProductFamily: strPtr("Data Payload"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/-DataScanned-Bytes/")},
				},
			},
		})
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
