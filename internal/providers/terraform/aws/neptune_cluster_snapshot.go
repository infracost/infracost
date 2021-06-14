package aws

import (
	"github.com/infracost/infracost/internal/schema"
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
	resourceData = dbClusterIdentifier[0]
	region := d.Get("region").String()
	costComponents := make([]*schema.CostComponent, 0)
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: backupCostComponent(resourceData, u, costComponents, region),
	}
}
