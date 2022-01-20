package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type EKSCluster struct {
	Address string
	Region  string
}

var EKSClusterUsageSchema = []*schema.UsageItem{}

func (r *EKSCluster) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *EKSCluster) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name:           r.Address,
		CostComponents: []*schema.CostComponent{r.clusterHoursCostComponent()},
		UsageSchema:    EKSClusterUsageSchema,
	}
}

func (r *EKSCluster) clusterHoursCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "EKS cluster",
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
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
