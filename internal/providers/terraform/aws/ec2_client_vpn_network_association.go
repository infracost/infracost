package aws

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetEC2ClientVPNNetworkAssociationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ec2_client_vpn_network_association",
		RFunc: NewEC2ClientVPNNetworkAssociation,
	}
}

func NewEC2ClientVPNNetworkAssociation(ctx *config.ProjectContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           "Endpoint association",
				Unit:           "hours",
				UnitMultiplier: decimal.NewFromInt(1),
				HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName: strPtr("aws"),
					Region:     strPtr(region),
					Service:    strPtr("AmazonVPC"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/ClientVPN-EndpointHours/")},
					},
				},
			},
		},
	}
}
