package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetNewEKSFargateProfileItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_eks_fargate_profile",
		RFunc: NewEKSFargateProfile,
	}
}

func NewEKSFargateProfile(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()
	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, memoryCostComponent(d, region))
	costComponents = append(costComponents, vcpuCostComponent(d, region))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func memoryCostComponent(d *schema.ResourceData, region string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Per GB per hour",
		Unit:           "GB-hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonEKS"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/Fargate-GB-Hours/")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

func vcpuCostComponent(d *schema.ResourceData, region string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Per vCPU per hour",
		Unit:           "CPU-hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonEKS"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/Fargate-vCPU-Hours:perCPU/")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}
