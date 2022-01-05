package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getKinesisDataAnalyticsRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_kinesisanalyticsv2_application",
		RFunc: NewKinesisanalyticsv2Application,
		Notes: []string{
			"Terraform doesn’t currently support Analytics Studio, but when it does they will require 2 orchestration KPUs.",
		},
	}
}
func NewKinesisanalyticsv2Application(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.Kinesisanalyticsv2Application{Address: strPtr(d.Address), RuntimeEnvironment: strPtr(d.Get("runtime_environment").String()), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
