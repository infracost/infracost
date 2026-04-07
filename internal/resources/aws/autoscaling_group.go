package aws

import (
	"context"
	"math"

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
	LaunchConfiguration *LaunchConfiguration
	LaunchTemplate      *LaunchTemplate
}

var AutoscalingGroupUsageSchema = append([]*schema.UsageItem{
	{Key: "instances", DefaultValue: 0, ValueType: schema.Int64},
}, InstanceUsageSchema...)

func (a *AutoscalingGroup) CoreType() string {
	return "AutoscalingGroup"
}

func (a *AutoscalingGroup) UsageSchema() []*schema.UsageItem {
	return a.getUsageSchemaWithDefaultInstanceCount()
}

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

// getUsageSchemaWithDefaultInstanceCount is a temporary hack to make --sync-usage-file use the group's "desired_size"
// as the default value for the "instances" usage param.  Without this, --sync-usage-file sets instances=0 causing the
// costs for the node group to be $0.  This can be removed when --sync-usage-file creates the usage file with usgage keys
// commented out by default.
func (a *AutoscalingGroup) getUsageSchemaWithDefaultInstanceCount() []*schema.UsageItem {
	var instanceCount *int64
	if a.LaunchConfiguration != nil {
		instanceCount = a.LaunchConfiguration.InstanceCount
	} else if a.LaunchTemplate != nil {
		instanceCount = a.LaunchTemplate.InstanceCount
	}

	if instanceCount == nil || *instanceCount == 0 {
		return AutoscalingGroupUsageSchema
	}

	usageSchema := make([]*schema.UsageItem, 0, len(AutoscalingGroupUsageSchema))
	for _, u := range AutoscalingGroupUsageSchema {
		if u.Key == "instances" {
			usageSchema = append(usageSchema, &schema.UsageItem{Key: "instances", DefaultValue: intVal(instanceCount), ValueType: schema.Int64})
		} else {
			usageSchema = append(usageSchema, u)
		}
	}
	return usageSchema
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

	estimate := func(ctx context.Context, u map[string]any) error {
		if estimateInstanceQualities != nil {
			err := estimateInstanceQualities(ctx, u)
			if err != nil {
				return err
			}
		}
		if a.Name != "" {
			count, err := aws.AutoscalingGetInstanceCount(ctx, a.Region, a.Name)
			if err != nil {
				return err
			}
			if count > 0 {
				u["instances"] = int64(math.Round(count))
			}
		}
		return nil
	}

	return &schema.Resource{
		Name:           a.Address,
		UsageSchema:    a.UsageSchema(),
		CostComponents: costComponents,
		SubResources:   subResources,
		EstimateUsage:  estimate,
	}
}
