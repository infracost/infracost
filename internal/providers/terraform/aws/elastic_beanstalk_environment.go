package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getElasticBeanstalkEnvironmentRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_elastic_beanstalk_environment",
		RFunc: newElasticBeanstalkEnvironment,
	}
}

func newElasticBeanstalkEnvironment(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {

	var loadBalancerType = ""
	var instanceType = ""
	var instanceCount int64
	var streamLogs bool
	var rdsIncluded bool
	var rootVolumeSize int64
	var rootVolumeType = "gp2"
	var rootVolumeIOPS int64

	autoscalingGroup := &schema.ResourceData{}
	data := autoscalingGroup

	for _, setting := range d.Get("setting").Array() {
		if setting.Get("name").String() == "InstanceTypes" {
			instanceType = setting.Get("value").String()
		}
		if setting.Get("name").String() == "MinSize" {
			instanceCount = setting.Get("value").Int()
		}
		if setting.Get("name").String() == "RootVolumeSize" {
			rootVolumeSize = setting.Get("value").Int()
		}
		if setting.Get("name").String() == "RootVolumeIOPS" {
			rootVolumeIOPS = setting.Get("value").Int()
		}
		if setting.Get("name").String() == "RootVolumeType" {
			rootVolumeType = setting.Get("value").String()
		}
		if setting.Get("name").String() == "LoadBalancerType" {
			data.Set("load_balancer_type", setting.Get("value").String())
			loadBalancerType = setting.Get("value").String()
		}
		if setting.Get("name").String() == "StreamLogs" {
			streamLogs = true
		}
		if setting.Get("name").String() == "DBInstanceClass" {
			rdsIncluded = true
			data.Set("db_instance_class", setting.Get("value").String())
		}
		if setting.Get("name").String() == "MultiAZDatabase" {
			data.Set("multi_az", true)
		}
		if setting.Get("name").String() == "DBEngine" {
			data.Set("engine", setting.Get("value").String())
		}
		if setting.Get("name").String() == "DBAllocatedStorage" {
			data.Set("allocated_storage", setting.Get("value").Float())
		}
	}
	data.Set("region", d.Get("region").String())
	data.Set("name", d.Get("name").String())

	r := &aws.ElasticBeanstalkEnvironment{
		Address:          d.Address,
		Region:           d.Get("region").String(),
		Name:             d.Get("name").String(),
		InstanceType:     instanceType,
		InstanceCount:    instanceCount,
		LoadBalancerType: loadBalancerType,
		StreamLogs:       streamLogs,
		RDSIncluded:      rdsIncluded,
		RootVolumeSize:   &rootVolumeSize,
		RootVolumeType:   rootVolumeType,
		RootVolumeIOPS:   rootVolumeIOPS,
	}

	if loadBalancerType == "classic" {
		r.ElasticLoadBalancer = newELB(data, u)
	} else {
		r.LoadBalancer = newLB(data, u)
	}

	if streamLogs {
		r.CloudwatchLogGroup = newCloudwatchLogGroup(data, u)
	}

	if rdsIncluded {
		r.DBInstance = newDBInstance(data, u)
	}

	r.PopulateUsage(u)

	return r.BuildResource()
}

func newLB(d *schema.ResourceData, u *schema.UsageData) *aws.LB {
	loadBalancerType := d.Get("load_balancer_type").String()

	r := &aws.LB{
		Address:          "aws_load_balancer",
		Region:           d.Get("region").String(),
		LoadBalancerType: loadBalancerType,
	}
	r.PopulateUsage(u)
	return r
}

func newELB(d *schema.ResourceData, u *schema.UsageData) *aws.ELB {
	r := &aws.ELB{
		Address: "aws_elb",
		Region:  d.Get("region").String(),
	}
	r.PopulateUsage(u)
	return r
}

func newCloudwatchLogGroup(d *schema.ResourceData, u *schema.UsageData) *aws.CloudwatchLogGroup {
	r := &aws.CloudwatchLogGroup{
		Address: "aws_cloudwatch_log_group",
		Region:  d.Get("region").String(),
	}

	r.PopulateUsage(u)
	return r
}

func newDBInstance(d *schema.ResourceData, u *schema.UsageData) *aws.DBInstance {

	r := &aws.DBInstance{
		Address:       "aws_db_instance",
		Region:        d.Get("region").String(),
		InstanceClass: d.Get("db_instance_class").String(),
		Engine:        d.Get("engine").String(),
		MultiAZ:       d.Get("multi_az").Bool(),
	}

	if !d.IsEmpty("allocated_storage") {
		r.AllocatedStorageGB = floatPtr(d.Get("allocated_storage").Float())
	}

	r.PopulateUsage(u)
	return r
}
