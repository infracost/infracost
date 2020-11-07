package aws

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetEIPRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_eip",
		RFunc: NewEIP,
	}
}

func NewEIP(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	// The IP address is probably used if it has an instance or network_interface, the instance might
	// be stopped but that's probably less likely
	if (d.Get("customer_owned_ipv4_pool").Exists() && d.Get("customer_owned_ipv4_pool").String() != "") ||
		d.Get("instance").Exists() || d.Get("network_interface").Exists() {
		return &schema.Resource{
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	region := d.Get("region").String()

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           "IP address (if unused)",
				Unit:           "hours",
				UnitMultiplier: 1,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonEC2"),
					ProductFamily: strPtr("IP Address"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/ElasticIP:IdleAddress/")},
					},
				},
				PriceFilter: &schema.PriceFilter{
					StartUsageAmount: strPtr("1"),
				},
			},
		},
	}
}
