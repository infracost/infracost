package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type EKSFargateProfile struct {
	Address string
	Region  string
}

func (r *EKSFargateProfile) CoreType() string {
	return "EKSFargateProfile"
}

func (r *EKSFargateProfile) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *EKSFargateProfile) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *EKSFargateProfile) BuildResource() *schema.Resource {
	costComponents := make([]*schema.CostComponent, 0)
	costComponents = append(costComponents, r.memoryCostComponent())
	costComponents = append(costComponents, r.vcpuCostComponent())

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *EKSFargateProfile) memoryCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Per GB per hour",
		Unit:           "GB",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
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

func (r *EKSFargateProfile) vcpuCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Per vCPU per hour",
		Unit:           "CPU",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
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
