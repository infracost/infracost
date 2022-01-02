package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getMQBrokerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_mq_broker",
		RFunc: NewMqBroker,
	}
}
func NewMqBroker(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.MqBroker{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String()), EngineType: strPtr(d.Get("engine_type").String()), HostInstanceType: strPtr(d.Get("host_instance_type").String())}
	if !d.IsEmpty("storage_type") {
		r.StorageType = strPtr(d.Get("storage_type").String())
	}
	if !d.IsEmpty("deployment_mode") {
		r.DeploymentMode = strPtr(d.Get("deployment_mode").String())
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
