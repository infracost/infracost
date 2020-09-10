package aws

import (
	"fmt"

	"github.com/infracost/infracost/pkg/schema"
	"github.com/shopspring/decimal"
)

func NewElasticsearchDomain(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {

	// domain_name := d.Get("domain_name").String()
	clusterConfig := d.Get("cluster_config")
	fmt.Printf("%v\n\n\n", clusterConfig)
	instanceType := clusterConfig.Array()[0].Get("instance_type").String()
	fmt.Printf("instanceType: %v", instanceType)
	region := d.Get("region").String()

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           "Per instance hour",
				Unit:           "hours",
				HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonES"),
					ProductFamily: strPtr("Elastic Search Instance"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/ESInstance/")},
						{Key: "instanceType", Value: &instanceType},
					},
				},
			},
		},
	}
}
