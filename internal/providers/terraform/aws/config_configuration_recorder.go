package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getConfigurationRecorderItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_config_configuration_recorder",
		RFunc: NewConfigConfigurationRecorder,
	}
}
func NewConfigConfigurationRecorder(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.ConfigConfigurationRecorder{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
