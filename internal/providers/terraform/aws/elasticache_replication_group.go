package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getElastiCacheReplicationGroupItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_elasticache_replication_group",
		CoreRFunc:           NewElastiCacheReplicationGroup,
		ReferenceAttributes: []string{"aws_appautoscaling_target.resource_id"},
		CustomRefIDFunc: func(d *schema.ResourceData) []string {
			// returns a name that will match the custom format used by aws_appautoscaling_target.resource_id
			name := d.Get("replication_group_id").String()
			if name != "" {
				return []string{"replication-group/" + name}
			}
			return nil
		},
	}
}
func NewElastiCacheReplicationGroup(d *schema.ResourceData) schema.CoreResource {
	cacheClusters := d.GetInt64OrDefault("num_cache_clusters", 1)
	if d.IsEmpty("num_cache_clusters") && !d.IsEmpty("number_cache_clusters") {
		// check for deprecated attribute
		cacheClusters = d.Get("number_cache_clusters").Int()
	}

	clusterNodeGroups := d.Get("num_node_groups").Int()
	if d.IsEmpty("num_node_groups") {
		// check for deprecated attribute
		clusterNodeGroups = d.Get("cluster_mode.0.num_node_groups").Int()
	}

	clusterReplicasPerNodeGroup := d.Get("replicas_per_node_group").Int()
	if d.IsEmpty("replicas_per_node_group") {
		// check for deprecated attribute
		clusterReplicasPerNodeGroup = d.Get("cluster_mode.0.replicas_per_node_group").Int()
	}

	targets := []*aws.AppAutoscalingTarget{}
	for _, ref := range d.References("aws_appautoscaling_target.resource_id") {
		targets = append(targets, newAppAutoscalingTarget(ref, ref.UsageData))
	}

	r := &aws.ElastiCacheReplicationGroup{
		Address:                     d.Address,
		Region:                      d.Get("region").String(),
		NodeType:                    d.Get("node_type").String(),
		Engine:                      d.Get("engine").String(),
		CacheClusters:               cacheClusters,
		ClusterNodeGroups:           clusterNodeGroups,
		ClusterReplicasPerNodeGroup: clusterReplicasPerNodeGroup,
		SnapshotRetentionLimit:      d.Get("snapshot_retention_limit").Int(),
		AppAutoscalingTarget:        targets,
	}
	return r
}
