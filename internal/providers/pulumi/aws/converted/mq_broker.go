package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getMQBrokerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_mq_broker",
		RFunc: NewMQBroker,
	}
}
func NewMQBroker(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.MQBroker{
		Address:          d.Address,
		Region:           d.Get("region").String(),
		EngineType:       d.Get("engineType").String(),
		HostInstanceType: d.Get("hostInstanceType").String(),
		StorageType:      d.Get("storageType").String(),
		DeploymentMode:   d.Get("deploymentMode").String(),
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
