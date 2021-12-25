package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type KinesisDataAnalyticsSnapshot struct {
	Address                    *string
	Region                     *string
	DurableApplicationBackupGb *int64 `infracost_usage:"durable_application_backup_gb"`
}

var KinesisDataAnalyticsSnapshotUsageSchema = []*schema.UsageItem{{Key: "durable_application_backup_gb", ValueType: schema.Int64, DefaultValue: 0}}

func (r *KinesisDataAnalyticsSnapshot) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *KinesisDataAnalyticsSnapshot) BuildResource() *schema.Resource {
	region := *r.Region
	costComponents := make([]*schema.CostComponent, 0)
	var durableApplicationBackupGb *decimal.Decimal

	if r != nil && r.DurableApplicationBackupGb != nil {
		durableApplicationBackupGb = decimalPtr(decimal.NewFromInt(*r.DurableApplicationBackupGb))
	}

	costComponents = append(costComponents, kinesisBackupCostComponent(region, durableApplicationBackupGb))

	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: KinesisDataAnalyticsSnapshotUsageSchema,
	}
}
