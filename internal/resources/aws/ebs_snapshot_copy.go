package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type EbsSnapshotCopy struct {
	Address       string
	Region        string
	VolumeRefSize float64
}

var EbsSnapshotCopyUsageSchema = []*schema.UsageItem{}

func (r *EbsSnapshotCopy) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *EbsSnapshotCopy) BuildResource() *schema.Resource {
	region := r.Region

	gbVal := decimal.NewFromInt(int64(defaultVolumeSize))

	if r.VolumeRefSize != 0 {
		gbVal = decimal.NewFromFloat(r.VolumeRefSize)
	}

	costComponents := []*schema.CostComponent{
		ebsSnapshotCostComponent(region, gbVal),
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents, UsageSchema: EbsSnapshotCopyUsageSchema,
	}
}
