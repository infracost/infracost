package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

type AutoscalingGroup struct {
	// "required" args that can't really be missing.
	Address  string
	Region   string
	Capacity int64

	// "optional" args, that may be empty depending on the resource config
	LaunchConfiguration *LaunchConfiguration
}

var AutoscalingGroupUsageSchema = []*schema.UsageSchemaItem{
	{Key: "capacity", DefaultValue: 0, ValueType: schema.Int64},
}

func (a *AutoscalingGroup) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(a, u)
}

func (a *AutoscalingGroup) BuildResource() *schema.Resource {
	costComponents := make([]*schema.CostComponent, 0)
	subResources := make([]*schema.Resource, 0)

	if a.LaunchConfiguration != nil {
		lc := a.LaunchConfiguration.BuildResource()
		schema.MultiplyQuantities(lc, decimal.NewFromInt(a.Capacity))
		subResources = append(subResources, lc)
	}

	return &schema.Resource{
		Name:           a.Address,
		UsageSchema:    AutoscalingGroupUsageSchema,
		CostComponents: costComponents,
		SubResources:   subResources,
	}
}
