package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type Kinesisanalyticsv2ApplicationSnapshot struct {
	Address                    *string
	Region                     *string
	DurableApplicationBackupGb *float64 `infracost_usage:"durable_application_backup_gb"`
}

var Kinesisanalyticsv2ApplicationSnapshotUsageSchema = []*schema.UsageItem{{Key: "durable_application_backup_gb", ValueType: schema.Float64, DefaultValue: 0}}

func (r *Kinesisanalyticsv2ApplicationSnapshot) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *Kinesisanalyticsv2ApplicationSnapshot) BuildResource() *schema.Resource {
	region := *r.Region
	costComponents := make([]*schema.CostComponent, 0)
	var durableApplicationBackupGb *decimal.Decimal

	if r.DurableApplicationBackupGb != nil {
		durableApplicationBackupGb = decimalPtr(decimal.NewFromFloat(*r.DurableApplicationBackupGb))
	}

	costComponents = append(costComponents, kinesisBackupCostComponent(region, durableApplicationBackupGb))

	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: Kinesisanalyticsv2ApplicationSnapshotUsageSchema,
	}
}
