package aws

import (
	"infracost/pkg/schema"

	"github.com/shopspring/decimal"
)

func NewEBSSnapshotCopy(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	gbVal := decimal.NewFromInt(int64(defaultVolumeSize))

	sourceSnapshotRefs := d.References("source_snapshot_id")
	if len(sourceSnapshotRefs) > 0 {
		volumeRefs := sourceSnapshotRefs[0].References("volume_id")
		if len(volumeRefs) > 0 {
			if volumeRefs[0].Get("size").Exists() {
				gbVal = decimal.NewFromFloat(volumeRefs[0].Get("size").Float())
			}
		}
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: ebsSnapshotCostComponents(region, gbVal),
	}
}
