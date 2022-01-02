package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getKinesisDataAnalyticsSnapshotRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_kinesisanalyticsv2_application_snapshot",
		RFunc: NewKinesisanalyticsv2ApplicationSnapshot,
	}
}
func NewKinesisanalyticsv2ApplicationSnapshot(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.Kinesisanalyticsv2ApplicationSnapshot{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
