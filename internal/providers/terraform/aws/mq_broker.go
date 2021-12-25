package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

func GetMQBrokerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_mq_broker",
		RFunc: NewMQBroker,
	}
}
func NewMQBroker(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.MQBroker{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String()), EngineType: strPtr(d.Get("engine_type").String()), HostInstanceType: strPtr(d.Get("host_instance_type").String())}
	if d.Get("storage_type").Exists() && d.Get("storage_type").Type != gjson.Null {
		r.StorageType = strPtr(d.Get("storage_type").String())
	}
	if d.Get("deployment_mode").Exists() && d.Get("deployment_mode").Type != gjson.Null {
		r.DeploymentMode = strPtr(d.Get("deployment_mode").String())
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
