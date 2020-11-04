package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetNewEKSNodeGroupItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_eks_node_group",
		RFunc: NewEKSNodeGroup,
	}
}

func NewEKSNodeGroup(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()
	scalingConfig := d.Get("scaling_config").Array()[0]
	desiredSize := decimal.NewFromInt(scalingConfig.Get("desired_size").Int())
	vcpuVal := desiredSize.Mul(decimal.NewFromInt(1))

	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, hoursCostComponent(d, region))
	costComponents = append(costComponents, vcpuCostComponent(d, vcpuVal, region))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func hoursCostComponent(d *schema.ResourceData, region string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "EKS Cluster memory charges",
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonEKS"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", Value: strPtr("USE1-AmazonEKS-Hours:perCluster")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

func vcpuCostComponent(d *schema.ResourceData, vcpuVal decimal.Decimal, region string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "EKS Cluster CPU charges",
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(vcpuVal),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonEKS"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", Value: strPtr("USE1-Fargate-vCPU-Hours:perCPU")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}
