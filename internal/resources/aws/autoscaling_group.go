package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

type AutoscalingGroup struct {
	// "required" args that can't really be missing.
	Address string
	Region  string

	// "optional" args, that may be empty depending on the resource config
	LaunchConfiguration *LaunchConfiguration
	LaunchTemplate      *LaunchTemplate
}

var AutoscalingGroupUsageSchema = append([]*schema.UsageSchemaItem{
	{Key: "instances", DefaultValue: 0, ValueType: schema.Int64},
}, InstanceUsageSchema...)

func (a *AutoscalingGroup) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(a, u)

	// The usage keys for Launch Template and Configuration are specified on the Autoscaling Group resource
	if a.LaunchTemplate != nil {
		resources.PopulateArgsWithUsage(a.LaunchTemplate, u)
	}

	if a.LaunchConfiguration != nil {
		resources.PopulateArgsWithUsage(a.LaunchConfiguration, u)
	}
}

func (a *AutoscalingGroup) BuildResource() *schema.Resource {
	costComponents := make([]*schema.CostComponent, 0)
	subResources := make([]*schema.Resource, 0)

	if a.LaunchConfiguration != nil {
		lc := a.LaunchConfiguration.BuildResource()
		// If the Launch Configuration returns nil it is not supported so the Autoscaling Group should also return nil
		if lc == nil {
			return nil
		}
		subResources = append(subResources, lc)
	} else if a.LaunchTemplate != nil {
		lt := a.LaunchTemplate.BuildResource()
		// If the Launch Template returns nil it is not supported so the Autoscaling Group should also return nil
		if lt == nil {
			return nil
		}
		subResources = append(subResources, lt)
	}

	return &schema.Resource{
		Name:           a.Address,
		UsageSchema:    AutoscalingGroupUsageSchema,
		CostComponents: costComponents,
		SubResources:   subResources,
	}
}
