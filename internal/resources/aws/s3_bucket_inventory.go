package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type S3BucketInventory struct {
	Address              string
	Region               string
	MonthlyListedObjects *int64 `infracost_usage:"monthly_listed_objects"`
}

var S3BucketInventoryUsageSchema = []*schema.UsageItem{{Key: "monthly_listed_objects", ValueType: schema.Int64, DefaultValue: 0}}

func (r *S3BucketInventory) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *S3BucketInventory) BuildResource() *schema.Resource {
	var listedObj *decimal.Decimal
	if r.MonthlyListedObjects != nil {
		listedObj = decimalPtr(decimal.NewFromInt(*r.MonthlyListedObjects))
	}

	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Objects listed",
				Unit:            "1M objects",
				UnitMultiplier:  decimal.NewFromInt(1000000),
				MonthlyQuantity: listedObj,
				ProductFilter: &schema.ProductFilter{
					VendorName: strPtr("aws"),
					Region:     strPtr(r.Region),
					Service:    strPtr("AmazonS3"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/Inventory-ObjectsListed/")},
					},
				},
			},
		},
		UsageSchema: S3BucketInventoryUsageSchema,
	}
}
