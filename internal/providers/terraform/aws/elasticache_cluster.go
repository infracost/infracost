package aws

import (
	"fmt"
	"github.com/infracost/infracost/internal/schema"
	"strings"

	"github.com/shopspring/decimal"
)

func GetElastiCacheClusterItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_elasticache_cluster",
		RFunc:               NewElastiCacheCluster,
		ReferenceAttributes: []string{"replication_group_id"},
	}
}

func NewElastiCacheCluster(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	var nodeType string
	var cacheNodes int64
	var cacheEngine string

	var snapShotRetentionLimit decimal.Decimal
	var backupRetention decimal.Decimal
	var monthlyBackupStorageTotal decimal.Decimal

	if d.Get("node_type").Exists() {
		cacheNodes = d.Get("num_cache_nodes").Int()
		nodeType = d.Get("node_type").String()
		cacheEngine = d.Get("engine").String()
	}

	if d.Get("snapshot_retention_limit").Exists() {
		snapShotRetentionLimit = decimal.NewFromInt(d.Get("snapshot_retention_limit").Int())
	}

	replicationGroupID := d.References("replication_group_id")

	if len(replicationGroupID) > 0 {
		return &schema.Resource{
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Elasticache (on-demand, %s)", nodeType),
			Unit:           "hours",
			UnitMultiplier: 1,
			HourlyQuantity: decimalPtr(decimal.NewFromInt(cacheNodes)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonElastiCache"),
				ProductFamily: strPtr("Cache Instance"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "instanceType", Value: strPtr(nodeType)},
					{Key: "locationType", Value: strPtr("AWS Region")},
					{Key: "cacheEngine", Value: strPtr(strings.Title(cacheEngine))},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
		},
	}

	if cacheEngine == "redis" && snapShotRetentionLimit.GreaterThan(decimal.NewFromInt(1)) {
		backupRetention = snapShotRetentionLimit.Sub(decimal.NewFromInt(1))

		if u != nil && u.Get("backup_storage").Exists() {
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
