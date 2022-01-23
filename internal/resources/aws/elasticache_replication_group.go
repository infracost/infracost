package aws

import (
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

type ElastiCacheReplicationGroup struct {
	Address                     string
	Region                      string
	NodeType                    string
	Engine                      string
	CacheClusters               int64
	ClusterDisabled             bool
	ClusterMode                 string
	ClusterNodeGroups           int64
	ClusterReplicasPerNodeGroup int64
	SnapshotRetentionLimit      int64
	SnapshotStorageSizeGB       *float64 `infracost_usage:"snapshot_storage_size_gb"`
}

var ElastiCacheReplicationGroupUsageSchema = []*schema.UsageItem{
	{Key: "snapshot_storage_size_gb", ValueType: schema.Float64, DefaultValue: 0},
}

func (r *ElastiCacheReplicationGroup) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ElastiCacheReplicationGroup) BuildResource() *schema.Resource {
	engine := r.Engine
	if engine == "" {
		engine = "redis"
	}

	cacheNodes := r.CacheClusters

	if r.ClusterMode != "" && !r.ClusterDisabled {
		cacheNodes = (r.ClusterNodeGroups * r.ClusterReplicasPerNodeGroup) + r.ClusterNodeGroups
	}

	cluster := &ElastiCacheCluster{
		Region:                 r.Region,
		NodeType:               r.NodeType,
		Engine:                 engine,
		CacheNodes:             cacheNodes,
		SnapshotRetentionLimit: r.SnapshotRetentionLimit,
		SnapshotStorageSizeGB:  r.SnapshotStorageSizeGB,
	}

	costComponents := []*schema.CostComponent{
		cluster.elastiCacheCostComponent(),
	}

	if strings.ToLower(engine) == "redis" && r.SnapshotRetentionLimit > 1 {
		costComponents = append(costComponents, cluster.backupStorageCostComponent())
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    ElastiCacheReplicationGroupUsageSchema,
	}
}
