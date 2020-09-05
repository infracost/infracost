package aws

import (
	"infracost/pkg/schema"

	"github.com/shopspring/decimal"
)

func NewLB(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	costComponentName := "Per Application Load Balancer"
	productFamily := "Load Balancer-Application"
	if d.Get("load_balancer_type").String() == "network" {
		costComponentName = "Per Network Load Balancer"
		productFamily = "Load Balancer-Network"
	}

	return newLBResource(d, u, productFamily, costComponentName)
}

func newLBResource(d *schema.ResourceData, u *schema.ResourceData, productFamily string, costComponentName string) *schema.Resource {
	region := d.Get("region").String()

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           costComponentName,
				Unit:           "hours",
				HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AWSELB"),
					ProductFamily: strPtr(productFamily),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/LoadBalancerUsage/")},
					},
				},
			},
		},
	}
}
