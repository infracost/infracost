package aws

import (
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetKinesisDataAnalyticsSnapshotRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_kinesisanalyticsv2_application_snapshot",
		RFunc:               NewKinesisDataAnalyticsSnapshot,
		ReferenceAttributes: []string{"application_name"},
	}
}

func NewKinesisDataAnalyticsSnapshot(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	costComponents := make([]*schema.CostComponent, 0)
	var durableApplicationBackupGb *decimal.Decimal

	applicationName := d.References("application_name")
	runtimeEnvironment := applicationName[0].Get("runtime_environment").String()

	if u != nil && u.Get("durable_application_backup_gb").Type != gjson.Null {
		durableApplicationBackupGb = decimalPtr(decimal.NewFromInt(u.Get("durable_application_backup_gb").Int()))
	}

	if strings.HasPrefix(strings.ToLower(runtimeEnvironment), "flink") {
		costComponents = append(costComponents, kinesisBackupCostComponent(region, durableApplicationBackupGb))
	} else {
		return &schema.Resource{
			NoPrice:   true,
			IsSkipped: true,
		}
	}
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
