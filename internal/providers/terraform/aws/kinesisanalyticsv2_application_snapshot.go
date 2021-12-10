package aws

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetKinesisDataAnalyticsSnapshotRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_kinesisanalyticsv2_application_snapshot",
		RFunc: NewKinesisDataAnalyticsSnapshot,
	}
}

func NewKinesisDataAnalyticsSnapshot(ctx *config.ProjectContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	costComponents := make([]*schema.CostComponent, 0)
	var durableApplicationBackupGb *decimal.Decimal

	if u != nil && u.Get("durable_application_backup_gb").Type != gjson.Null {
		durableApplicationBackupGb = decimalPtr(decimal.NewFromInt(u.Get("durable_application_backup_gb").Int()))
	}

	costComponents = append(costComponents, kinesisBackupCostComponent(region, durableApplicationBackupGb))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
