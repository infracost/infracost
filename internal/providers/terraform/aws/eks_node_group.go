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
	desiredSize := decimal.NewFromInt(d.Get("desired_size").Int()) // TODO in block
	vcpuVal := desiredSize.Mul(decimal.NewFromInt(1))              // TODO find flavor
	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, hoursCostComponent(d))
	costComponents = append(costComponents, vcpuCostComponent(d, vcpuVal))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func hoursCostComponent(d *schema.ResourceData) *schema.CostComponent {
	region := d.Get("region").String()
	return &schema.CostComponent{
		Name:           "EKS Cluster",
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

func vcpuCostComponent(d *schema.ResourceData, vcpuVal decimal.Decimal) *schema.CostComponent {
	region := d.Get("region").String()
	return &schema.CostComponent{
		Name:           "EKS Cluster",
		Unit:           "vCPU-hours",
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
