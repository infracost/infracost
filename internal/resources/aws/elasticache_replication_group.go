package aws

import (
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

type ElastiCacheReplicationGroup struct {
	Address                       string
	Region                        string
	NodeType                      string
	Engine                        string
	CacheClusters                 int64
	ClusterNodeGroups             int64
	ClusterReplicasPerNodeGroup   int64
	SnapshotRetentionLimit        int64
	SnapshotStorageSizeGB         *float64 `infracost_usage:"snapshot_storage_size_gb"`
	ReservedInstanceTerm          *string  `infracost_usage:"reserved_instance_term"`
	ReservedInstancePaymentOption *string  `infracost_usage:"reserved_instance_payment_option"`

	AppAutoscalingTarget []*AppAutoscalingTarget
}

func (r *ElastiCacheReplicationGroup) CoreType() string {
	return "ElastiCacheReplicationGroup"
}

func (r *ElastiCacheReplicationGroup) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "snapshot_storage_size_gb", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "reserved_instance_term", DefaultValue: "", ValueType: schema.String},
		{Key: "reserved_instance_payment_option", DefaultValue: "", ValueType: schema.String},
	}
}

func (r *ElastiCacheReplicationGroup) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ElastiCacheReplicationGroup) BuildResource() *schema.Resource {
	engine := r.Engine
	if engine == "" {
		engine = "redis"
	}

	var autoscaling bool
	nodeGroups := r.ClusterNodeGroups
	replicasPerNodeGroup := r.ClusterReplicasPerNodeGroup
	for _, target := range r.AppAutoscalingTarget {
		switch target.ScalableDimension {
		case "elasticache:replication-group:NodeGroups":
			autoscaling = true
			if target.Capacity != nil {
				nodeGroups = *target.Capacity
			} else {
				nodeGroups = target.MinCapacity
			}
		case "elasticache:replication-group:Replicas":
			autoscaling = true
			if target.Capacity != nil {
				replicasPerNodeGroup = *target.Capacity
			} else {
				replicasPerNodeGroup = target.MinCapacity
			}
		}
	}

	cacheNodes := r.CacheClusters
	if nodeGroups > 0 {
		// CacheClusters is mutually exclusive with ClusterNodeGroups/ClusterReplicasPerNodeGroup
		cacheNodes = (nodeGroups * replicasPerNodeGroup) + nodeGroups
	}

	cluster := &ElastiCacheCluster{
		Region:                        r.Region,
		NodeType:                      r.NodeType,
		Engine:                        engine,
		CacheNodes:                    cacheNodes,
		SnapshotRetentionLimit:        r.SnapshotRetentionLimit,
		SnapshotStorageSizeGB:         r.SnapshotStorageSizeGB,
		ReservedInstanceTerm:          r.ReservedInstanceTerm,
		ReservedInstancePaymentOption: r.ReservedInstancePaymentOption,
	}

	costComponents := []*schema.CostComponent{
		cluster.elastiCacheCostComponent(autoscaling),
	}

	if strings.ToLower(engine) == "redis" && r.SnapshotRetentionLimit > 1 {
		costComponents = append(costComponents, cluster.backupStorageCostComponent())
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}
