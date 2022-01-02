package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type EksCluster struct {
	Address *string
	Region  *string
}

var EksClusterUsageSchema = []*schema.UsageItem{}

func (r *EksCluster) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *EksCluster) BuildResource() *schema.Resource {
	region := *r.Region

	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, clusterHoursCostComponent(r, region))

	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: EksClusterUsageSchema,
	}
}

func clusterHoursCostComponent(r *EksCluster, region string) *schema.CostComponent {
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
