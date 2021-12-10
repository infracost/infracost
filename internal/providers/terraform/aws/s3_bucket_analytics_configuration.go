package aws

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetS3BucketAnalyticsConfigurationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_s3_bucket_analytics_configuration",
		RFunc: NewS3BucketAnalyticsConfiguration,
	}
}

func NewS3BucketAnalyticsConfiguration(ctx *config.ProjectContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	var monitObj *decimal.Decimal
	if u != nil && u.Get("monthly_monitored_objects").Exists() {
		monitObj = decimalPtr(decimal.NewFromInt(u.Get("monthly_monitored_objects").Int()))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Objects monitored",
				Unit:            "1M objects",
				UnitMultiplier:  decimal.NewFromInt(1000000),
				MonthlyQuantity: monitObj,
				ProductFilter: &schema.ProductFilter{
					VendorName: strPtr("aws"),
					Region:     strPtr(region),
					Service:    strPtr("AmazonS3"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/StorageAnalytics-ObjCount/")},
					},
				},
			},
		},
	}
}
