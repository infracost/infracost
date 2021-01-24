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

	instanceCount := decimal.Zero

	var instanceTier string

	if u != nil && u.Get("instance_tier").Exists() {
		instanceTier = u.Get("instance_tier").String()
	} else if d.Get("registration_limit").Exists() {
		if d.Get("registration_limit").Int() > 1000 {
			instanceTier = "Advanced"
		}
	}

	if u != nil && u.Get("instance_count").Exists() {
		instanceCount = decimal.NewFromInt(u.Get("instance_count").Int())
	}

	switch instanceTier {
	case "Advanced":
		return &schema.Resource{
			Name: d.Address,
			CostComponents: []*schema.CostComponent{
				{
					Name:           "On-prem managed instances (advanced)",
					Unit:           "hours",
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
	default:
		return nil
	}
}
