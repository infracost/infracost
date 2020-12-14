package aws

import (
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetECRRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ecr_repository",
		RFunc: NewECRRepository,
	}
}

func NewECRRepository(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	var storageSize *decimal.Decimal

	if u != nil && u.Get("storage_size.0.value").Exists() {
		storageSize = decimalPtr(decimal.NewFromFloat(u.Get("storage_size.0.value").Float()))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Storage",
				Unit:            "GB-months",
				UnitMultiplier:  1,
				MonthlyQuantity: storageSize,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonECR"),
					ProductFamily: strPtr("EC2 Container Registry"),
				},
			},
		},
	}
}
