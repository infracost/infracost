package aws

import (
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetSSMActivationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ssm_activation",
		RFunc: NewSSMActivation,
	}
}

func NewSSMActivation(ctx *config.ProjectContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	var instanceCount *decimal.Decimal
	var instanceTier string

	if u != nil && u.Get("instance_tier").Exists() {
		instanceTier = u.Get("instance_tier").String()
	} else if d.Get("registration_limit").Exists() {
		if d.Get("registration_limit").Int() > 1000 {
			instanceTier = "Advanced"
		}
	}

	if u != nil && u.Get("instances").Exists() {
		instanceCount = decimalPtr(decimal.NewFromInt(u.Get("instances").Int()))
	}

	if strings.ToLower(instanceTier) == "advanced" {
		return &schema.Resource{
			Name: d.Address,
			CostComponents: []*schema.CostComponent{
				{
					Name:           "On-prem managed instances (advanced)",
					Unit:           "hours",
					UnitMultiplier: decimal.NewFromInt(1),
					HourlyQuantity: instanceCount,
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

	// standard instanceTier is free
	return &schema.Resource{
		Name:      d.Address,
		NoPrice:   true,
		IsSkipped: true,
	}
}
