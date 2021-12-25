package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetConfigurationRecorderItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_config_configuration_recorder",
		RFunc: NewConfigurationRecorderItem,
	}
}
func NewConfigurationRecorderItem(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.ConfigurationRecorderItem{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
