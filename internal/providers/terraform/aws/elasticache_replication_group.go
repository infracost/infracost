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

	var snapShotRetentionLimit decimal.Decimal
	var backupRetention decimal.Decimal
	var monthlyBackupStorageTotal decimal.Decimal

	if d.Get("engine").Exists() {
		cacheEngine = d.Get("engine").String()
	} else {
		cacheEngine = "redis"
	}

	if d.Get("snapshot_retention_limit").Exists() {
		snapShotRetentionLimit = decimal.NewFromInt(d.Get("snapshot_retention_limit").Int())
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

	if cacheEngine == "redis" && snapShotRetentionLimit.GreaterThan(decimal.NewFromInt(1)) {
		backupRetention = snapShotRetentionLimit.Sub(decimal.NewFromInt(1))

		if u != nil && u.Get("monthly_backup_storage").Exists() {
			snapshotSize := decimal.NewFromInt(u.Get("snapshot_storage_size.0.value").Int())
			monthlyBackupStorageTotal = snapshotSize.Mul(backupRetention)
		}

		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Backup storage",
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
