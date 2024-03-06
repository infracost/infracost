package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type DocDBClusterSnapshot struct {
	Address         string
	Region          string
	BackupStorageGB *float64 `infracost_usage:"backup_storage_gb"`
}

func (r *DocDBClusterSnapshot) CoreType() string {
	return "DocDBClusterSnapshot"
}

func (r *DocDBClusterSnapshot) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "backup_storage_gb", ValueType: schema.Float64, DefaultValue: 0},
	}
}

func (r *DocDBClusterSnapshot) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *DocDBClusterSnapshot) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	var backupStorage *decimal.Decimal
	if r.BackupStorageGB != nil {
		backupStorage = decimalPtr(decimal.NewFromFloat(*r.BackupStorageGB))
	}

	cluster := &DocDBCluster{
		Region: r.Region,
	}

	costComponents = append(costComponents, cluster.backupStorageCostComponent(backupStorage))

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}
