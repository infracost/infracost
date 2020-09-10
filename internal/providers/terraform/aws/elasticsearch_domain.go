package aws

import (
	"fmt"

	"github.com/infracost/infracost/pkg/schema"
	"github.com/shopspring/decimal"
)

func NewElasticsearchDomain(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {

	domainName := d.Get("domain_name").String()
	region := d.Get("region").String()
	clusterConfig := d.Get("cluster_config").Array()[0]
	instanceType := clusterConfig.Get("instance_type").String()
	instanceCount := clusterConfig.Get("instance_count").Int()
	ebsOptions := d.Get("ebs_options").Array()[0]

	ebsTypeMap := map[string]*string{
		"gp2": strPtr("GP2"),
		"io1": strPtr("PIOPS"),
		"io2": strPtr("PIOPS"),
		"st1": strPtr("Magnetic"),
		"sc1": strPtr("Magnetic"),
	}

	gbVal := decimal.NewFromInt(int64(defaultVolumeSize))
	if ebsOptions.Get("volume_size").Exists() {
		gbVal = decimal.NewFromFloat(ebsOptions.Get("volume_size").Float())
	}

	ebsType := "gp2"
	if ebsOptions.Get("volume_type").Exists() {
		ebsType = ebsOptions.Get("volume_type").String()
	}

	fmt.Printf(*ebsTypeMap[ebsType])

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Per instance(x%d) hour (%s)", instanceCount, domainName),
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
		{
			Name:           "Storage",
			Unit:           "GB-months",
			HourlyQuantity: &gbVal,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonES"),
				ProductFamily: strPtr("Elastic Search Volume"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/ES.+-Storage/")},
					{Key: "storageMedia", Value: strPtr("GP2")},
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
