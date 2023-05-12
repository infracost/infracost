package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetConfigurationRecorderItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "states.aws.config.config_recorder.present",
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
