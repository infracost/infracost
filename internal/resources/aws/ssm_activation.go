package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"strings"

	"github.com/shopspring/decimal"
)

type SSMActivation struct {
	Address           string
	Region            string
	RegistrationLimit int64
	InstanceTier      *string `infracost_usage:"instance_tier"`
	Instances         *int64  `infracost_usage:"instances"`
}

func (r *SSMActivation) CoreType() string {
	return "SSMActivation"
}

func (r *SSMActivation) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "instance_tier", ValueType: schema.String, DefaultValue: "standard"},
		{Key: "instances", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *SSMActivation) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *SSMActivation) BuildResource() *schema.Resource {
	var instanceTier string
	if r.InstanceTier != nil {
		instanceTier = *r.InstanceTier
	} else if r.RegistrationLimit > 1000 {
		instanceTier = "Advanced"
	}

	var instanceCount *decimal.Decimal
	if r.Instances != nil {
		instanceCount = decimalPtr(decimal.NewFromInt(*r.Instances))
	}

	if strings.ToLower(instanceTier) == "advanced" {
		return &schema.Resource{
			Name: r.Address,
			CostComponents: []*schema.CostComponent{
				{
					Name:           "On-prem managed instances (advanced)",
					Unit:           "hours",
					UnitMultiplier: decimal.NewFromInt(1),
					HourlyQuantity: instanceCount,
					ProductFilter: &schema.ProductFilter{
						VendorName:    strPtr("aws"),
						Region:        strPtr(r.Region),
						Service:       strPtr("AWSSystemsManager"),
						ProductFamily: strPtr("AWS Systems Manager"),
						AttributeFilters: []*schema.AttributeFilter{
							{Key: "usagetype", ValueRegex: strPtr("/MI-AdvInstances-Hrs/")},
						},
					},
					UsageBased: true,
				},
			}, UsageSchema: r.UsageSchema(),
		}
	}

	return &schema.Resource{
		Name:        r.Address,
		NoPrice:     true,
		IsSkipped:   true,
		UsageSchema: r.UsageSchema(),
	}
}
