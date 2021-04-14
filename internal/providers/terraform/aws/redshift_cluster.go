package aws

import (
	"fmt"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetRedshiftClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_redshift_cluster",
		RFunc: NewRedshiftCluster,
	}
}

func NewRedshiftCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	nodeType := d.Get("node_type").String()
	numberOfNodes := int64(1)
	if d.Get("number_of_nodes").Type != gjson.Null {
		numberOfNodes = d.Get("number_of_nodes").Int()
	}

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Redshift Cluster (%s)", nodeType),
			Unit:           "hours",
			UnitMultiplier: 1,
			HourlyQuantity: decimalPtr(decimal.NewFromInt(numberOfNodes)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonRedshift"),
				ProductFamily: strPtr("Compute Instance"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "instanceType", Value: strPtr(nodeType)},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
		},
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
