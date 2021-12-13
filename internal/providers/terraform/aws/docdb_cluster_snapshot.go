package aws

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetDocDBClusterSnapshotRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_docdb_cluster_snapshot",
		RFunc: NewDocDBClusterSnapshot,
	}

}

func NewDocDBClusterSnapshot(ctx *config.RunContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	costComponents := []*schema.CostComponent{}

	var backupStorage *decimal.Decimal
	if u != nil && u.Get("backup_storage_gb").Exists() {
		backupStorage = decimalPtr(decimal.NewFromInt(u.Get("backup_storage_gb").Int()))
		costComponents = append(costComponents, docDBCluster(region, backupStorage))
	} else {

		var unknown *decimal.Decimal

		costComponents = append(costComponents, docDBCluster(region, unknown))

	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
