package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getMQBrokerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_mq_broker",
		CoreRFunc: NewMQBroker,
	}
}
func NewMQBroker(d *schema.ResourceData) schema.CoreResource {
	r := &aws.MQBroker{
		Address:          d.Address,
		Region:           d.Get("region").String(),
		EngineType:       d.Get("engine_type").String(),
		HostInstanceType: d.Get("host_instance_type").String(),
		StorageType:      d.Get("storage_type").String(),
		DeploymentMode:   d.Get("deployment_mode").String(),
	}
	return r
}
