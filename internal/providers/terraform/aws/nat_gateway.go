package aws

import (
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetNATGatewayRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_nat_gateway",
		RFunc: NewNATGateway,
	}
}

func NewNATGateway(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	var gbDataProcessed *decimal.Decimal
	if u != nil && u.Get("monthly_gb_data_processed").Exists() {
		gbDataProcessed = decimalPtr(decimal.NewFromFloat(u.Get("monthly_gb_data_processed").Float()))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           "NAT gateway",
				Unit:           "hours",
				UnitMultiplier: 1,
				HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonEC2"),
					ProductFamily: strPtr("NAT Gateway"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/NatGateway-Hours/")},
					},
				},
			},
			{
				Name:            "Data processed",
				Unit:            "GB",
				UnitMultiplier:  1,
				MonthlyQuantity: gbDataProcessed,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonEC2"),
					ProductFamily: strPtr("NAT Gateway"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/NatGateway-Bytes/")},
					},
				},
			},
		},
	}
}
