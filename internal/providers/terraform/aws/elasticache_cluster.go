package aws

import (
	"fmt"
	"github.com/infracost/infracost/internal/schema"
	"strings"

	"github.com/shopspring/decimal"
)

func GetElastiCacheClusterItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_elasticache_cluster",
		RFunc: NewElastiCacheCluster,
	}
}

func NewElastiCacheCluster(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	var nodeType string
	var cacheNodes int64
	var cacheEngine string

	snapShotRetentionLimit := decimal.Zero
	backupRetention := decimal.Zero
	monthlyBackupStorageTotal := decimal.Zero

	if d.Get("node_type").Exists() {
		cacheNodes = d.Get("num_cache_nodes").Int()
		nodeType = d.Get("node_type").String()
		cacheEngine = d.Get("engine").String()
	}

	if cacheEngine == "redis" && d.Get("snapshot_retention_limit").Exists() {
		snapShotRetentionLimit = decimal.NewFromInt(d.Get("snapshot_retention_limit").Int())
		backupRetention = snapShotRetentionLimit.Sub(decimal.NewFromInt(1))
	}

	if u != nil && u.Get("monthly_backup_storage").Exists() {
		snapshotSize := decimal.NewFromInt(u.Get("snapshot_storage_size.0.value").Int())
		monthlyBackupStorageTotal = snapshotSize.Mul(backupRetention)
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
					{Key: "cacheEngine", Value: strPtr(strings.Title(cacheEngine))},
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
