package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getElastiCacheReplicationGroupItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_elasticache_replication_group",
		RFunc:           NewElastiCacheReplicationGroup,
		ReferenceAttributes: []string{"awsAppautoscalingTarget.resourceId"},
		CustomRefIDFunc: func(d *schema.ResourceData) []string {
			// returns a name that will match the custom format used by aws_appautoscaling_target.resource_id
			name := d.Get("replicationGroupId").String()
			if name != "" {
				return []string{"replication-group/" + name}
			}
			return nil
		},
	}
}
func NewElastiCacheReplicationGroup(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	cacheClusters := d.GetInt64OrDefault("numCacheClusters", 1)
	if d.IsEmpty("num_cache_clusters") && !d.IsEmpty("number_cache_clusters") {
		// check for deprecated attribute
		cacheClusters = d.Get("numberCacheClusters").Int()
	}

	clusterNodeGroups := d.Get("numNodeGroups").Int()
	if d.IsEmpty("num_node_groups") {
		// check for deprecated attribute
		clusterNodeGroups = d.Get("clusterMode.0.numNodeGroups").Int()
	}

	clusterReplicasPerNodeGroup := d.Get("replicasPerNodeGroup").Int()
	if d.IsEmpty("replicas_per_node_group") {
		// check for deprecated attribute
		clusterReplicasPerNodeGroup = d.Get("clusterMode.0.replicasPerNodeGroup").Int()
	}

	targets := []*aws.AppAutoscalingTarget{}
	for _, ref := range d.References("awsAppautoscalingTarget.resourceId") {
		targets = append(targets, newAppAutoscalingTarget(ref, ref.UsageData))
	}

	r := &aws.ElastiCacheReplicationGroup{
		Address:                     d.Address,
		Region:                      d.Get("region").String(),
		NodeType:                    d.Get("nodeType").String(),
		Engine:                      d.Get("engine").String(),
		CacheClusters:               cacheClusters,
		ClusterNodeGroups:           clusterNodeGroups,
		ClusterReplicasPerNodeGroup: clusterReplicasPerNodeGroup,
		SnapshotRetentionLimit:      d.Get("snapshotRetentionLimit").Int(),
		AppAutoscalingTarget:        targets,
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
