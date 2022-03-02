package aws

import (
	"context"
	"math"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage/aws"
)

// ElasticBeanstalkEnvironment struct represents <TODO: cloud service short description>.
//
// <TODO: Add any important information about the resource and links to the
// pricing pages or documentation that might be useful to developers in the future, e.g:>
//
// Resource information: https://aws.amazon.com/elasticbeanstalk/
// Pricing information: https://aws.amazon.com/elasticbeanstalk/pricing/
type ElasticBeanstalkEnvironment struct {
	Address string
	Region  string
	Name    string

	InstanceCount  int64
	InstanceType   string
	RootVolumeIOPS int64
	RootVolumeSize *int64
	RootVolumeType string

	RDSIncluded bool

	LoadBalancerType string
	AutoScalingGroup *AutoscalingGroup

	StreamLogs            bool
	instanceCount         int64
	MonthlyDataIngestedGB *float64 `infracost_usage:"monthly_data_ingested_gb"`
	StorageGB             *float64 `infracost_usage:"storage_gb"`
	MonthlyDataScannedGB  *float64 `infracost_usage:"monthly_data_scanned_gb"`
	CloudwatchLogGroup    *CloudwatchLogGroup
	LoadBalancer          *LB
	ElasticLoadBalancer   *ELB
	DBInstance            *DBInstance
}

// ElasticBeanstalkEnvironmentUsageSchema defines a list which represents the usage schema of ElasticBeanstalkEnvironment.
var ElasticBeanstalkEnvironmentUsageSchema = append([]*schema.UsageItem{
	{Key: "monthly_data_processed_gb", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "new_connections", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "active_connections", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "rule_evaluations", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "processed_bytes_gb", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "monthly_data_ingested_gb", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "monthly_data_scanned_gb", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "storage_gb", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "instances", DefaultValue: 0, ValueType: schema.Int64},
}, InstanceUsageSchema...)

// PopulateUsage parses the u schema.UsageData into the ElasticBeanstalkEnvironment.
// It uses the `infracost_usage` struct tags to populate data into the ElasticBeanstalkEnvironment.
func (r *ElasticBeanstalkEnvironment) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
	if r.LoadBalancerType == "classic" {
		resources.PopulateArgsWithUsage(r.ElasticLoadBalancer, u)
	} else {
		resources.PopulateArgsWithUsage(r.LoadBalancer, u)
	}
	if r.StreamLogs {
		resources.PopulateArgsWithUsage(r.CloudwatchLogGroup, u)
	}
	if r.RDSIncluded {
		resources.PopulateArgsWithUsage(r.DBInstance, u)
	}
}

// getUsageSchemaWithDefaultInstanceCount is a temporary hack to make --sync-usage-file use the group's "desired_size"
// as the default value for the "instances" usage param.  Without this, --sync-usage-file sets instances=0 causing the
// costs for the node group to be $0.  This can be removed when --sync-usage-file creates the usage file with usgage keys
// commented out by default.
func (r *ElasticBeanstalkEnvironment) getUsageSchemaWithDefaultInstanceCount() []*schema.UsageItem {
	var instanceCount = &r.instanceCount

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

// BuildResource builds a schema.Resource from a valid ElasticBeanstalkEnvironment struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *ElasticBeanstalkEnvironment) BuildResource() *schema.Resource {

	subResources := make([]*schema.Resource, 0)

	rootBlockDevice := &EBSVolume{
		Address: "aws_ebs_volume",
		Region:  r.Region,
		Type:    r.RootVolumeType,
		Size:    r.RootVolumeSize,
		IOPS:    r.RootVolumeIOPS,
	}

	launchConfiguration := &LaunchConfiguration{
		Address:         "aws_launch_configuration",
		Region:          r.Region,
		InstanceType:    r.InstanceType,
		InstanceCount:   &r.InstanceCount,
		RootBlockDevice: rootBlockDevice,
	}

	autoScalingGroup := &AutoscalingGroup{
		Address:             "aws_autoscaling_group",
		Region:              r.Region,
		Name:                r.Name,
		LaunchConfiguration: launchConfiguration,
	}

	autoScalingGroupResource := autoScalingGroup.BuildResource()

	var estimateInstanceQualities = autoScalingGroupResource.EstimateUsage

	subResources = append(subResources, autoScalingGroupResource.SubResources...)
	if r.LoadBalancerType == "classic" {
		subResources = append(subResources, r.ElasticLoadBalancer.BuildResource())
	} else {
		subResources = append(subResources, r.LoadBalancer.BuildResource())
	}

	if r.StreamLogs {
		subResources = append(subResources, r.CloudwatchLogGroup.BuildResource())
	}

	if r.RDSIncluded {
		subResources = append(subResources, r.DBInstance.BuildResource())
	}

	estimate := func(ctx context.Context, u map[string]interface{}) error {
		if estimateInstanceQualities != nil {
			err := estimateInstanceQualities(ctx, u)
			if err != nil {
				return err
			}
		}
		if r.Name != "" {
			count, err := aws.AutoscalingGetInstanceCount(ctx, r.Region, r.Name)
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
		Name:           r.Address,
		UsageSchema:    r.getUsageSchemaWithDefaultInstanceCount(),
		CostComponents: autoScalingGroupResource.CostComponents,
		SubResources:   subResources,
		EstimateUsage:  estimate,
	}

}
