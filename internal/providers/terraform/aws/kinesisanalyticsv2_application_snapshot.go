package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetKinesisDataAnalyticsSnapshotRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_kinesisanalyticsv2_application_snapshot",
		RFunc: NewKinesisDataAnalyticsSnapshot,
	}
}
func NewKinesisDataAnalyticsSnapshot(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.KinesisDataAnalyticsSnapshot{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
