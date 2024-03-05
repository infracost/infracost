package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getKinesisAnalyticsV2ApplicationSnapshotRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_kinesisanalyticsv2_application_snapshot",
		CoreRFunc: NewKinesisAnalyticsV2ApplicationSnapshot,
	}
}

func NewKinesisAnalyticsV2ApplicationSnapshot(d *schema.ResourceData) schema.CoreResource {
	r := &aws.KinesisAnalyticsV2ApplicationSnapshot{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}
	return r
}
