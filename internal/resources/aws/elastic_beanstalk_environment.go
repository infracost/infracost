package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
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

	InstanceCount    *int64
	InstanceType     string
	RDSIncluded      bool
	LoadBalancerType string

	StreamLogs            bool
	MonthlyDataIngestedGB *float64 `infracost_usage:"cloudwatch_monthly_data_ingested_gb"`
	StorageGB             *float64 `infracost_usage:"cloudwatch_storage_gb"`
	MonthlyDataScannedGB  *float64 `infracost_usage:"cloudwatch_monthly_data_scanned_gb"`
	MonthlyDataProcessed  *float64 `infracost_usage:"monthly_data_processed_gb"`
	LBNewConnections      *float64 `infracost_usage:"lb_new_connections"`
	LBActiveConnections   *float64 `infracost_usage:"lb_active_connections"`
	LBRuleEvaluations     *float64 `infracost_usage:"lb_rule_evaluations"`
	LBProcessedBytesGB    *float64 `infracost_usage:"lb_processed_bytes_gb"`

	RootBlockDevice     *EBSVolume
	CloudwatchLogGroup  *CloudwatchLogGroup
	LoadBalancer        *LB
	ElasticLoadBalancer *ELB
	DBInstance          *DBInstance
	LaunchConfiguration *LaunchConfiguration
}

// ElasticBeanstalkEnvironmentUsageSchema defines a list which represents the usage schema of ElasticBeanstalkEnvironment.
var ElasticBeanstalkEnvironmentUsageSchema = append([]*schema.UsageItem{
	{Key: "monthly_data_processed_gb", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "lb_new_connections", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "lb_active_connections", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "lb_rule_evaluations", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "lb_processed_bytes_gb", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "cloudwatch_monthly_data_ingested_gb", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "cloudwatch_monthly_data_scanned_gb", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "cloudwatch_storage_gb", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "instances", DefaultValue: 0, ValueType: schema.Int64},
}, LaunchConfigurationUsageSchema...)

// PopulateUsage parses the u schema.UsageData into the ElasticBeanstalkEnvironment.
// It uses the `infracost_usage` struct tags to populate data into the ElasticBeanstalkEnvironment.
func (r *ElasticBeanstalkEnvironment) PopulateUsage(u *schema.UsageData) {
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
	resources.PopulateArgsWithUsage(r.LaunchConfiguration, u)
}

// BuildResource builds a schema.Resource from a valid ElasticBeanstalkEnvironment struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *ElasticBeanstalkEnvironment) BuildResource() *schema.Resource {

	a := &schema.Resource{
		Name:        r.Address,
		UsageSchema: ElasticBeanstalkEnvironmentUsageSchema,
	}

	a.SubResources = append(a.SubResources, r.LaunchConfiguration.BuildResource())
	a.EstimateUsage = r.LaunchConfiguration.BuildResource().EstimateUsage
	if r.RDSIncluded {
		dbInstance := r.DBInstance.BuildResource()
		a.SubResources = append(a.SubResources, dbInstance)

	}
	if r.StreamLogs {
		a.SubResources = append(a.SubResources, r.CloudwatchLogGroup.BuildResource())
	}
	switch r.LoadBalancerType {
	case "classic":
		a.SubResources = append(a.SubResources, r.ElasticLoadBalancer.BuildResource())

	default:
		a.SubResources = append(a.SubResources, r.LoadBalancer.BuildResource())
	}

	return a

}
