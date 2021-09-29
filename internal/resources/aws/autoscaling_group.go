package aws

import (
	"context"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage/aws"
)

type AutoscalingGroup struct {
	// "required" args that can't really be missing.
	Address string
	Region  string
	Name    string

	// "optional" args, that may be empty depending on the resource config
	DesiredCapacity     *int64
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
	var estimateInstanceQualities schema.EstimateFunc

	if a.LaunchConfiguration != nil {
		lc := a.LaunchConfiguration.BuildResource()
		// If the Launch Configuration returns nil it is not supported so the Autoscaling Group should also return nil
		if lc == nil {
			return nil
		}
		subResources = append(subResources, lc)
		estimateInstanceQualities = lc.EstimateUsage
	} else if a.LaunchTemplate != nil {
		lt := a.LaunchTemplate.BuildResource()
		// If the Launch Template returns nil it is not supported so the Autoscaling Group should also return nil
		if lt == nil {
			return nil
		}
		subResources = append(subResources, lt)
		estimateInstanceQualities = lt.EstimateUsage
	}

	estimate := func(ctx context.Context, u map[string]interface{}) error {
		if a.DesiredCapacity != nil {
			// as a default
			u["instances"] = *a.DesiredCapacity
		}
		if a.Name != "" {
			// actual usage overrides desired capacity
			count, err := aws.AutoscalingGetInstanceCount(ctx, a.Region, a.Name)
			if err != nil {
				return err
			}
			if count > 0 {
				u["instances"] = count
			}
		}
		err := estimateInstanceQualities(ctx, u)
		if err != nil {
			return err
		}
		return nil
	}

	return &schema.Resource{
		Name:           a.Address,
		UsageSchema:    AutoscalingGroupUsageSchema,
		CostComponents: costComponents,
		SubResources:   subResources,
		EstimateUsage:  estimate,
	}
}
