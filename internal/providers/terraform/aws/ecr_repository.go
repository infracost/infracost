package aws

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetECRRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ecr_repository",
		RFunc: NewECRRepository,
	}
}

func NewECRRepository(ctx *config.ProjectContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	var storageSize *decimal.Decimal

	if u != nil && u.Get("storage_gb").Exists() {
		storageSize = decimalPtr(decimal.NewFromFloat(u.Get("storage_gb").Float()))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Storage",
				Unit:            "GB",
				UnitMultiplier:  decimal.NewFromInt(1),
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
