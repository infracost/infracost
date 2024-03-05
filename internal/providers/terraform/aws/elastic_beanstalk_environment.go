package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getElasticBeanstalkEnvironmentRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_elastic_beanstalk_environment",
		CoreRFunc: newElasticBeanstalkEnvironment,
	}
}

func newElasticBeanstalkEnvironment(d *schema.ResourceData) schema.CoreResource {
	var region = d.Get("region").String()
	var streamLogs = false
	var dBIncluded = false

	r := &aws.ElasticBeanstalkEnvironment{
		Address: d.Address,
		Region:  region,
		Name:    d.Get("name").String(),
	}

	lc := &aws.LaunchConfiguration{
		Address: "aws_launch_configuration",
		Region:  region,
	}

	volume := &aws.EBSVolume{
		Address: "aws_ebs_volume",
		Region:  region,
	}

	db := &aws.DBInstance{
		Address: "aws_db_instance",
		Region:  region,
	}

	cwlg := &aws.CloudwatchLogGroup{
		Address: "aws_cloudwatch_log_group",
		Region:  region,
	}

	elb := &aws.ELB{
		Address: "aws_elb",
		Region:  region,
	}

	lb := &aws.LB{
		Address: "aws_loadbalancer",
		Region:  region,
	}

	for _, setting := range d.Get("setting").Array() {
		switch setting.Get("name").String() {
		case "InstanceTypes":
			lc.InstanceType = setting.Get("value").String()
		case "InstanceType":
			// InstanceType is deprecated, so we only use it if InstanceTypes is not set
			if lc.InstanceType == "" {
				lc.InstanceType = setting.Get("value").String()
			}
		case "MinSize":
			lc.InstanceCount = intPtr(setting.Get("value").Int())
		case "RootVolumeSize":
			volume.Size = intPtr(setting.Get("value").Int())
		case "RootVolumeIOPS":
			volume.IOPS = setting.Get("value").Int()
		case "RootVolumeType":
			volume.Type = setting.Get("value").String()
		case "LoadBalancerType":
			r.LoadBalancerType = setting.Get("value").String()
		case "StreamLogs":
			streamLogs = true
		case "DBInstanceClass":
			dBIncluded = true
			db.InstanceClass = setting.Get("value").String()
		case "MultiAZDatabase":
			db.MultiAZ = true
		case "DBEngine":
			db.Engine = setting.Get("value").String()
		case "DBAllocatedStorage":
			db.AllocatedStorageGB = floatPtr(setting.Get("value").Float())
		}
	}
	r.LaunchConfiguration = lc
	r.LaunchConfiguration.RootBlockDevice = volume

	if dBIncluded {
		r.DBInstance = db
	}

	if streamLogs {
		r.CloudwatchLogGroup = cwlg
	}

	if r.LoadBalancerType == "classic" {
		r.ElasticLoadBalancer = elb
	} else {
		r.LoadBalancer = lb
	}

	return r
}
