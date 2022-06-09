package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getEBSSnapshotCopyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ebs_snapshot_copy",
		RFunc: NewEBSSnapshotCopy,
		ReferenceAttributes: []string{
			"volumeId",
			"sourceSnapshotId",
		},
	}
}
func NewEBSSnapshotCopy(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.EBSSnapshotCopy{Address: d.Address, Region: d.Get("region").String()}
	sourceSnapshotRefs := d.References("sourceSnapshotId")
	if len(sourceSnapshotRefs) > 0 {
		volumeRefs := sourceSnapshotRefs[0].References("volumeId")
		if len(volumeRefs) > 0 {
			if volumeRefs[0].Get("size").Exists() {
				r.SizeGB = floatPtr(volumeRefs[0].Get("size").Float())
			}
		}
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
