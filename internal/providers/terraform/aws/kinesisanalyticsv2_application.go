package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getKinesisAnalyticsV2ApplicationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_kinesisanalyticsv2_application",
		RFunc: NewKinesisAnalyticsV2Application,
		Notes: []string{
			"Terraform doesn’t currently support Analytics Studio, but when it does they will require 2 orchestration KPUs.",
		},
	}
}

func NewKinesisAnalyticsV2Application(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.KinesisAnalyticsV2Application{
		Address:            d.Address,
		Region:             d.Get("region").String(),
		RuntimeEnvironment: d.Get("runtime_environment").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
