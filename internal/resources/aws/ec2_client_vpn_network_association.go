package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type Ec2ClientVpnNetworkAssociation struct {
	Address *string
	Region  *string
}

var Ec2ClientVpnNetworkAssociationUsageSchema = []*schema.UsageItem{}

func (r *Ec2ClientVpnNetworkAssociation) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *Ec2ClientVpnNetworkAssociation) BuildResource() *schema.Resource {
	region := *r.Region

	return &schema.Resource{
		Name: *r.Address,
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
		}, UsageSchema: Ec2ClientVpnNetworkAssociationUsageSchema,
	}
}
