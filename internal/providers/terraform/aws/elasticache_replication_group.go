package aws

import (
	"fmt"
	"github.com/infracost/infracost/internal/schema"
	"strings"

	"github.com/shopspring/decimal"
)

func GetElastiCacheReplicationGroupItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_elasticache_replication_group",
		RFunc: NewElastiCacheReplicationGroup,
	}
}

func NewElastiCacheReplicationGroup(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	nodeType := d.Get("node_type").String()

	var cacheNodes decimal.Decimal
	var cacheEngine string

	snapShotRetentionLimit := decimal.Zero
	backupRetention := decimal.Zero
	monthlyBackupStorageTotal := decimal.Zero

	if d.Get("engine").Exists() {
		cacheEngine = d.Get("engine").String()
	} else {
		cacheEngine = "redis"
	}

	if cacheEngine == "redis" && d.Get("snapshot_retention_limit").Exists() {
		snapShotRetentionLimit = decimal.NewFromInt(d.Get("snapshot_retention_limit").Int())
		backupRetention = snapShotRetentionLimit.Sub(decimal.NewFromInt(1))
	}

	if u != nil && u.Get("monthly_backup_storage").Exists() {
		snapshotSize := decimal.NewFromInt(u.Get("snapshot_storage_size.0.value").Int())
		monthlyBackupStorageTotal = snapshotSize.Mul(backupRetention)
	}

	if d.Get("cluster_mode").Exists() {
		nodeGroups := decimal.NewFromInt(d.Get("cluster_mode.0.num_node_groups").Int())
		shards := decimal.NewFromInt(d.Get("cluster_mode.0.replicas_per_node_group").Int())
		cacheNodes = nodeGroups.Mul(shards).Add(nodeGroups)
	}

	if d.Get("number_cache_clusters").Exists() {
		cacheNodes = decimal.NewFromInt(d.Get("number_cache_clusters").Int())
	}

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Elasticache (on-demand, %s)", nodeType),
			Unit:           "hours",
			UnitMultiplier: 1,
			HourlyQuantity: &cacheNodes,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonElastiCache"),
				ProductFamily: strPtr("Cache Instance"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "instanceType", Value: strPtr(nodeType)},
					{Key: "cacheEngine", Value: strPtr(strings.Title(cacheEngine))},
					{Key: "locationType", Value: strPtr("AWS Region")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
		},
	}

	if snapShotRetentionLimit.GreaterThan(decimal.NewFromInt(0)) {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Elasticache snapshot storage",
			Unit:            "GB-months",
			UnitMultiplier:  1,
			MonthlyQuantity: &monthlyBackupStorageTotal,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonElastiCache"),
				ProductFamily: strPtr("Storage Snapshot"),
			},
		})
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
