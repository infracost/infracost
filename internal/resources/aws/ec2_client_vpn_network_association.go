package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type EC2ClientVPNNetworkAssociation struct {
	Address string
	Region  string
}

func (r *EC2ClientVPNNetworkAssociation) CoreType() string {
	return "EC2ClientVPNNetworkAssociation"
}

func (r *EC2ClientVPNNetworkAssociation) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *EC2ClientVPNNetworkAssociation) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *EC2ClientVPNNetworkAssociation) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           "Endpoint association",
				Unit:           "hours",
				UnitMultiplier: decimal.NewFromInt(1),
				HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName: strPtr("aws"),
					Region:     strPtr(r.Region),
					Service:    strPtr("AmazonVPC"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/ClientVPN-EndpointHours/")},
					},
				},
			},
		},
		UsageSchema: r.UsageSchema(),
	}
}
