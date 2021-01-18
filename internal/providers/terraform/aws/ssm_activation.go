package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetSSMActivationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ssm_activation",
		RFunc: NewSSMActivation,
	}
}

func NewSSMActivation(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	tier := "Standard"
	instanceCount := decimal.Zero

	if u != nil && u.Get("instance_tier").Exists() {
		tier = u.Get("instance_tier").String()
	}

	if u != nil && u.Get("instance_count").Exists() {
		instanceCount = decimal.NewFromInt(u.Get("instance_count").Int())
	}

	if tier == "Standard" {
		return &schema.Resource{
			Name: d.Address,
			CostComponents: []*schema.CostComponent{
				{
					Name:           "On-Premises instance management - standard",
					Unit:           "Hours",
					UnitMultiplier: 1,
					HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				},
			},
		}
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           "On-Premises instance management - advanced",
				Unit:           "Hours",
				UnitMultiplier: 1,
				HourlyQuantity: decimalPtr(instanceCount),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AWSSystemsManager"),
					ProductFamily: strPtr("AWS Systems Manager"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/MI-AdvInstances-Hrs/")},
					},
				},
			},
		},
	}
}
