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
	var nodeType = ""
	var nodes = int64(0)
	var loadBalancerType = ""
	var streamLogs = false
	purchaseOption := "on_demand"

	for _, setting := range d.Get("setting").Array() {
		if setting.Get("name").String() == "InstanceTypes" {
			nodeType = setting.Get("value").String()
		}
		if setting.Get("name").String() == "MinSize" {
			nodes = setting.Get("value").Int()
		}
		if setting.Get("name").String() == "LoadBalancerType" {
			loadBalancerType = setting.Get("value").String()
		}
		if setting.Get("name").String() == "StreamLogs" {
			streamLogs = setting.Get("value").Bool()
		}
	}

	region := d.Get("region").String()
	r := &aws.ElasticBeanstalkEnvironment{
		LoadBalancerType: loadBalancerType,
		Address:          d.Address,
		Region:           region,
		InstanceType:     nodeType,
		Nodes:            nodes,
		PurchaseOption:   purchaseOption,
		StreamLogs:       streamLogs,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
