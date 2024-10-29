package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type S3BucketAnalyticsConfiguration struct {
	Address                 string
	Region                  string
	MonthlyMonitoredObjects *int64 `infracost_usage:"monthly_monitored_objects"`
}

func (r *S3BucketAnalyticsConfiguration) CoreType() string {
	return "S3BucketAnalyticsConfiguration"
}

func (r *S3BucketAnalyticsConfiguration) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_monitored_objects", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *S3BucketAnalyticsConfiguration) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *S3BucketAnalyticsConfiguration) BuildResource() *schema.Resource {
	var monitObj *decimal.Decimal
	if r.MonthlyMonitoredObjects != nil {
		monitObj = decimalPtr(decimal.NewFromInt(*r.MonthlyMonitoredObjects))
	}

	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Objects monitored",
				Unit:            "1M objects",
				UnitMultiplier:  decimal.NewFromInt(1000000),
				MonthlyQuantity: monitObj,
				ProductFilter: &schema.ProductFilter{
					VendorName: strPtr("aws"),
					Region:     strPtr(r.Region),
					Service:    strPtr("AmazonS3"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/StorageAnalytics-ObjCount/")},
					},
				},
				UsageBased: true,
			},
		},
		UsageSchema: r.UsageSchema(),
	}
}
