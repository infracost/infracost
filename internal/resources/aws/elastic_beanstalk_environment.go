package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
)

// ElasticBeanstalkEnvironment struct represents AWS Elastic Beanstalk environments.
//
// Resource information: https://aws.amazon.com/elasticbeanstalk/
// Pricing information: https://aws.amazon.com/elasticbeanstalk/pricing/
type ElasticBeanstalkEnvironment struct {
	Address string
	Region  string
	Name    string

	LoadBalancerType string

	RootBlockDevice     *EBSVolume
	CloudwatchLogGroup  *CloudwatchLogGroup
	LoadBalancer        *LB
	ElasticLoadBalancer *ELB
	DBInstance          *DBInstance
	LaunchConfiguration *LaunchConfiguration
}

func (r *ElasticBeanstalkEnvironment) CoreType() string {
	return "ElasticBeanstalkEnvironment"
}

// UsageSchema defines a list which represents the usage schema of ElasticBeanstalkEnvironment.
// Usage costs for Elastic Beanstalk come from sub resources as it is a wrapper for other AWS services.
func (r *ElasticBeanstalkEnvironment) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{
			Key:          "cloudwatch",
			DefaultValue: &usage.ResourceUsage{Name: "cloudwatch", Items: CloudwatchLogGroupUsageSchema},
			ValueType:    schema.SubResourceUsage,
		},
		{
			Key:          "lb",
			DefaultValue: &usage.ResourceUsage{Name: "lb", Items: LBUsageSchema},
			ValueType:    schema.SubResourceUsage,
		},
		{
			Key:          "elb",
			DefaultValue: &usage.ResourceUsage{Name: "elb", Items: ELBUsageSchema},
			ValueType:    schema.SubResourceUsage,
		},
		{
			Key:          "db",
			DefaultValue: &usage.ResourceUsage{Name: "db", Items: DBInstanceUsageSchema},
			ValueType:    schema.SubResourceUsage,
		},
		{
			Key:          "ec2",
			DefaultValue: &usage.ResourceUsage{Name: "ec2", Items: LaunchConfigurationUsageSchema},
			ValueType:    schema.SubResourceUsage,
		},
	}
}

// PopulateUsage parses the u schema.UsageData into the ElasticBeanstalkEnvironment.
// It uses the `infracost_usage` struct tags to populate data into the ElasticBeanstalkEnvironment.
func (r *ElasticBeanstalkEnvironment) PopulateUsage(u *schema.UsageData) {
	if u == nil {
		return
	}

	if r.ElasticLoadBalancer != nil {
		resources.PopulateArgsWithUsage(r.ElasticLoadBalancer, schema.NewUsageData("elb", u.Get("elb").Map()))
	}

	if r.LoadBalancer != nil {
		resources.PopulateArgsWithUsage(r.LoadBalancer, schema.NewUsageData("lb", u.Get("lb").Map()))
	}

	if r.DBInstance != nil {
		resources.PopulateArgsWithUsage(r.DBInstance, schema.NewUsageData("db", u.Get("db").Map()))
	}

	if r.CloudwatchLogGroup != nil {
		resources.PopulateArgsWithUsage(r.CloudwatchLogGroup, schema.NewUsageData("cloudwatch", u.Get("cloudwatch").Map()))
	}

	resources.PopulateArgsWithUsage(r.LaunchConfiguration, schema.NewUsageData("ec2", u.Get("ec2").Map()))
}

// BuildResource builds a schema.Resource from a valid ElasticBeanstalkEnvironment struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *ElasticBeanstalkEnvironment) BuildResource() *schema.Resource {
	a := &schema.Resource{
		Name:        r.Address,
		UsageSchema: r.UsageSchema(),
	}

	a.SubResources = append(a.SubResources, r.LaunchConfiguration.BuildResource())

	if r.DBInstance != nil {
		a.SubResources = append(a.SubResources, r.DBInstance.BuildResource())
	}

	if r.CloudwatchLogGroup != nil {
		a.SubResources = append(a.SubResources, r.CloudwatchLogGroup.BuildResource())
	}

	if r.LoadBalancerType == "classic" {
		a.SubResources = append(a.SubResources, r.ElasticLoadBalancer.BuildResource())
	} else {
		a.SubResources = append(a.SubResources, r.LoadBalancer.BuildResource())
	}

	return a

}
