package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type NewEKSFargateProfileItem struct {
	Address *string
	Region  *string
}

var NewEKSFargateProfileItemUsageSchema = []*schema.UsageItem{}

func (r *NewEKSFargateProfileItem) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *NewEKSFargateProfileItem) BuildResource() *schema.Resource {
	region := *r.Region
	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, memoryCostComponent(r, region))
	costComponents = append(costComponents, vcpuCostComponent(r, region))

	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: NewEKSFargateProfileItemUsageSchema,
	}
}

func memoryCostComponent(r *NewEKSFargateProfileItem, region string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Per GB per hour",
		Unit:           "GB",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
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

func vcpuCostComponent(r *NewEKSFargateProfileItem, region string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Per vCPU per hour",
		Unit:           "CPU",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
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
