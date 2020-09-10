package aws

import (
	"fmt"

	"github.com/infracost/infracost/pkg/schema"
	"github.com/shopspring/decimal"
)

func NewElasticsearchDomain(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {

	domainName := d.Get("domain_name").String()
	clusterConfig := d.Get("cluster_config").Array()[0]
	instanceType := clusterConfig.Get("instance_type").String()
	instanceCount := clusterConfig.Get("instance_count").Int()
	region := d.Get("region").String()

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           fmt.Sprintf("Per instance hour (%s)", domainName),
				Unit:           "hours",
				HourlyQuantity: decimalPtr(decimal.NewFromInt(instanceCount)),
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
				PriceFilter: &schema.PriceFilter{
					PurchaseOption: strPtr("on_demand"),
				},
			},
		},
	}
}
