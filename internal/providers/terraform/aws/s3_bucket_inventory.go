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

func NewS3BucketInventory(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Objects listed",
				Unit:            "objects",
				MonthlyQuantity: decimalPtr(decimal.NewFromInt(0)),
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
