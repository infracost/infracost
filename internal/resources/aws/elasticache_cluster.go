package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type ElastiCacheCluster struct {
	Address                string
	Region                 string
	HasReplicationGroup    bool
	NodeType               string
	Engine                 string
	CacheNodes             int64
	SnapshotRetentionLimit int64
	SnapshotStorageSizeGB  *float64 `infracost_usage:"snapshot_storage_size_gb"`
}

var ElastiCacheClusterUsageSchema = []*schema.UsageItem{
	{Key: "snapshot_storage_size_gb", ValueType: schema.Float64, DefaultValue: 0},
}

func (r *ElastiCacheCluster) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ElastiCacheCluster) BuildResource() *schema.Resource {
	if r.HasReplicationGroup {
		return &schema.Resource{
			Name:        r.Address,
			NoPrice:     true,
			IsSkipped:   true,
			UsageSchema: ElastiCacheClusterUsageSchema,
		}
	}

	costComponents := []*schema.CostComponent{
		r.elastiCacheCostComponent(),
	}

	if strings.ToLower(r.Engine) == "redis" && r.SnapshotRetentionLimit > 1 {
		costComponents = append(costComponents, r.backupStorageCostComponent())
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
	}
}

func (r *ElastiCacheCluster) elastiCacheCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           fmt.Sprintf("ElastiCache (on-demand, %s)", r.NodeType),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(r.CacheNodes)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonElastiCache"),
			ProductFamily: strPtr("Cache Instance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "instanceType", Value: strPtr(r.NodeType)},
				{Key: "locationType", Value: strPtr("AWS Region")},
				{Key: "cacheEngine", Value: strPtr(strings.Title(r.Engine))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

func (r *ElastiCacheCluster) backupStorageCostComponent() *schema.CostComponent {
	var monthlyBackupStorageGB *decimal.Decimal

	backupRetention := r.SnapshotRetentionLimit - 1

	if r.SnapshotStorageSizeGB != nil {
		snapshotSize := decimal.NewFromFloat(*r.SnapshotStorageSizeGB)
		monthlyBackupStorageGB = decimalPtr(snapshotSize.Mul(decimal.NewFromInt(backupRetention)))
	}

	return &schema.CostComponent{
		Name:            "Backup storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: monthlyBackupStorageGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonElastiCache"),
			ProductFamily: strPtr("Storage Snapshot"),
		},
	}
}
