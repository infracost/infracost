package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetNewKMSKeyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_kms_key",
		RFunc: NewKMSKey,
	}
}

func NewKMSKey(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {

	region := d.Get("region").String()

	costComponents := []*schema.CostComponent{
		{
			Name:            "Customer master key",
			Unit:            "months",
			UnitMultiplier:  1,
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("awskms"),
				ProductFamily: strPtr("Encryption Key"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "locationType", Value: strPtr("AWS Region")},
					{Key: "usagetype", ValueRegex: strPtr("/KMS-Keys/")},
				},
			},
		},
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
