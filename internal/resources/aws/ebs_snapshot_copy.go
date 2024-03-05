package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type EBSSnapshotCopy struct {
	Address string
	Region  string
	SizeGB  *float64
}

func (r *EBSSnapshotCopy) CoreType() string {
	return "EBSSnapshotCopy"
}

func (r *EBSSnapshotCopy) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *EBSSnapshotCopy) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *EBSSnapshotCopy) BuildResource() *schema.Resource {
	region := r.Region

	gbVal := decimal.NewFromInt(int64(defaultVolumeSize))

	if r.SizeGB != nil {
		gbVal = decimal.NewFromFloat(*r.SizeGB)
	}

	costComponents := []*schema.CostComponent{
		ebsSnapshotCostComponent(region, gbVal),
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}
