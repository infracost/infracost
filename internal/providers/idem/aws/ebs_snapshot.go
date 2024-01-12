package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetEBSSnapshotRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "states.aws.ec2.snapshot.present",
		RFunc:               NewEBSSnapshot,
		ReferenceAttributes: []string{"states.aws.ec2.volume.present:resource_id"},
	}
}
func NewEBSSnapshot(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.EBSSnapshot{Address: d.Address, Region: d.Get("region").String()}
	volumeRefs := d.References("states.aws.ec2.volume.present:resource_id")
	if len(volumeRefs) > 0 {
		if volumeRefs[0].Get("size").Exists() {
			r.SizeGB = floatPtr(volumeRefs[0].Get("size").Float())
		}
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
