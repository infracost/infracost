package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getEBSSnapshotRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_ebs_snapshot",
		RFunc:               NewEBSSnapshot,
		ReferenceAttributes: []string{"volume_id"},
	}
}
func NewEBSSnapshot(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.EBSSnapshot{Address: d.Address, Region: d.Get("region").String()}
	volumeRefs := d.References("volume_id")
	if len(volumeRefs) > 0 {
		if volumeRefs[0].Get("size").Exists() {
			r.SizeGB = floatPtr(volumeRefs[0].Get("size").Float())
		}
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
