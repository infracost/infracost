package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetKinesisDataAnalyticsRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_kinesisanalyticsv2_application",
		RFunc: NewKinesisDataAnalytics,
		Notes: []string{
			"Terraform doesn’t currently support Analytics Studio, but when it does they will require 2 orchestration KPUs.",
		},
	}
}
func NewKinesisDataAnalytics(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.KinesisDataAnalytics{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String()), RuntimeEnvironment: strPtr(d.Get("runtime_environment").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
