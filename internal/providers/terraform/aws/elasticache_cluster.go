package aws

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetElastiCacheClusterItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_elasticache_cluster",
		RFunc:               NewElastiCacheCluster,
		ReferenceAttributes: []string{"replication_group_id"},
	}
}

func NewElastiCacheCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var nodeType, cacheEngine string
	var cacheNodes decimal.Decimal

	replicationGroupID := d.References("replication_group_id")
	// If replicationGroupID is set, show costs in aws_elasticache_replication_group and not in this resource
	if len(replicationGroupID) > 0 {
		return &schema.Resource{
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	nodeType = d.Get("node_type").String()
	cacheEngine = d.Get("engine").String()
	cacheNodes = decimal.NewFromInt(d.Get("num_cache_nodes").Int())
	return newElasticacheResource(d, u, nodeType, cacheNodes, cacheEngine)
}

func newElasticacheResource(d *schema.ResourceData, u *schema.UsageData, nodeType string, cacheNodes decimal.Decimal, cacheEngine string) *schema.Resource {
	region := d.Get("region").String()
	var backupRetention, snapShotRetentionLimit decimal.Decimal

	if d.Get("snapshot_retention_limit").Exists() {
		snapShotRetentionLimit = decimal.NewFromInt(d.Get("snapshot_retention_limit").Int())
	}

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Elasticache (on-demand, %s)", nodeType),
			Unit:           "hours",
			UnitMultiplier: 1,
			HourlyQuantity: decimalPtr(cacheNodes),
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
		var monthlyBackupStorageTotal *decimal.Decimal

		if u != nil && u.Get("snapshot_storage_size_gb").Exists() {
			snapshotSize := decimal.NewFromInt(u.Get("snapshot_storage_size_gb").Int())
			monthlyBackupStorageTotal = decimalPtr(snapshotSize.Mul(backupRetention))
		}

		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Backup storage",
			Unit:            "GB-months",
			UnitMultiplier:  1,
			MonthlyQuantity: monthlyBackupStorageTotal,
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
