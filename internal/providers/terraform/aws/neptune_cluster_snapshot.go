package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetNeptuneClusterSnapshotRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_neptune_cluster_snapshot",
		RFunc: NewNeptuneClusterSnapshot,
		ReferenceAttributes: []string{
			"db_cluster_identifier",
		},
	}
}

func NewNeptuneClusterSnapshot(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var resourceData *schema.ResourceData
	dbClusterIdentifier := d.References("db_cluster_identifier")
	var retentionPeriod *decimal.Decimal

	if len(dbClusterIdentifier) > 0 {
		resourceData = dbClusterIdentifier[0]
		if resourceData.Get("backup_retention_period").Type != gjson.Null {
			retentionPeriod = decimalPtr(decimal.NewFromInt(resourceData.Get("backup_retention_period").Int()))
			if retentionPeriod.LessThan(decimal.NewFromInt(2)) {
				return &schema.Resource{
					Name:      d.Address,
					NoPrice:   true,
					IsSkipped: true,
				}
			}
		}
	}
	region := d.Get("region").String()

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: []*schema.CostComponent{backupCostComponent(u, region)},
	}
}
