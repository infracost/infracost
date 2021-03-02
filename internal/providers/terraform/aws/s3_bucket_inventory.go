package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetS3BucketInventoryRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_s3_bucket_inventory",
		RFunc: NewS3BucketInventory,
	}
}

func NewS3BucketInventory(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	var listedObj *decimal.Decimal
	if u != nil && u.Get("monthly_listed_objects").Exists() {
		listedObj = decimalPtr(decimal.NewFromInt(u.Get("monthly_listed_objects").Int()))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Objects listed",
				Unit:            "objects",
				UnitMultiplier:  1000000,
				MonthlyQuantity: listedObj,
				ProductFilter: &schema.ProductFilter{
					VendorName: strPtr("aws"),
					Region:     strPtr(region),
					Service:    strPtr("AmazonS3"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/Inventory-ObjectsListed/")},
					},
				},
			},
		},
	}
}
