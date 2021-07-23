package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetNewEKSClusterItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_eks_cluster",
		RFunc: NewEKSCluster,
	}
}

func NewEKSCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, clusterHoursCostComponent(d, region))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func clusterHoursCostComponent(d *schema.ResourceData, region string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "EKS cluster",
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonEKS"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/AmazonEKS-Hours:perCluster/")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}
