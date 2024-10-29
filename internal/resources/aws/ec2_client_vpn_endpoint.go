package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type EC2ClientVPNEndpoint struct {
	Address string
	Region  string
}

func (r *EC2ClientVPNEndpoint) CoreType() string {
	return "EC2ClientVPNEndpoint"
}

func (r *EC2ClientVPNEndpoint) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *EC2ClientVPNEndpoint) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *EC2ClientVPNEndpoint) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           "Connection",
				Unit:           "hours",
				UnitMultiplier: decimal.NewFromInt(1),
				HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName: strPtr("aws"),
					Region:     strPtr(r.Region),
					Service:    strPtr("AmazonVPC"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/ClientVPN-ConnectionHours/")},
					},
				},
			},
		},
		UsageSchema: r.UsageSchema(),
	}
}
