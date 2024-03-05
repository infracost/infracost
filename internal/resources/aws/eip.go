package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type EIP struct {
	Address   string
	Region    string
	Allocated bool
}

func (r *EIP) CoreType() string {
	return "EIP"
}

func (r *EIP) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *EIP) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *EIP) BuildResource() *schema.Resource {
	// The EIP is free if allocated. AWS does this to encourage efficient use of Elastic IPs
	// and discourage users from leaving unused EIPs lying around in their AWS account.
	if r.Allocated {
		return &schema.Resource{
			Name:        r.Address,
			NoPrice:     true,
			IsSkipped:   true,
			UsageSchema: r.UsageSchema(),
		}
	}

	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           "IP address (if unused)",
				Unit:           "hours",
				UnitMultiplier: decimal.NewFromInt(1),
				HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(r.Region),
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
		UsageSchema: r.UsageSchema(),
	}
}
