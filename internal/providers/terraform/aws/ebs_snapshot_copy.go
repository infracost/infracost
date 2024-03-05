package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getEBSSnapshotCopyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_ebs_snapshot_copy",
		CoreRFunc: NewEBSSnapshotCopy,
		ReferenceAttributes: []string{
			"volume_id",
			"source_snapshot_id",
		},
	}
}
func NewEBSSnapshotCopy(d *schema.ResourceData) schema.CoreResource {
	r := &aws.EBSSnapshotCopy{Address: d.Address, Region: d.Get("region").String()}
	sourceSnapshotRefs := d.References("source_snapshot_id")
	if len(sourceSnapshotRefs) > 0 {
		volumeRefs := sourceSnapshotRefs[0].References("volume_id")
		if len(volumeRefs) > 0 {
			if volumeRefs[0].Get("size").Exists() {
				r.SizeGB = floatPtr(volumeRefs[0].Get("size").Float())
			}
		}
	}
	return r
}
