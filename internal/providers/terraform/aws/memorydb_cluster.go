package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getMemoryDBClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_memorydb_cluster",
		CoreRFunc:           NewMemoryDBCluster,
		ReferenceAttributes: []string{"aws_appautoscaling_target.resource_id"},
		CustomRefIDFunc: func(d *schema.ResourceData) []string {
			// returns a name that will match the custom format used by aws_appautoscaling_target.resource_id
			name := d.Get("name").String()
			if name != "" {
				return []string{"memorydb:" + name}
			}
			return nil
		},
	}
}

func NewMemoryDBCluster(d *schema.ResourceData) schema.CoreResource {
	// Get the number of shards and replicas per shard
	numShards := d.Get("num_shards").Int()
	replicasPerShard := d.Get("num_replicas_per_shard").Int()

	// Get the engine type (redis or valkey)
	engine := d.Get("engine").String()
	if engine == "" {
		engine = "redis" // Default engine is Redis
	}

	// Get the snapshot retention limit
	snapshotRetentionLimit := d.Get("snapshot_retention_limit").Int()

	// Get autoscaling targets
	targets := []*aws.AppAutoscalingTarget{}
	for _, ref := range d.References("aws_appautoscaling_target.resource_id") {
		targets = append(targets, newAppAutoscalingTarget(ref, ref.UsageData))
	}

	// Create a MemoryDBCluster resource
	r := &aws.MemoryDBCluster{
		Address:                d.Address,
		Region:                 d.Get("region").String(),
		NodeType:               d.Get("node_type").String(),
		Engine:                 engine,
		NumShards:              numShards,
		ReplicasPerShard:       replicasPerShard,
		SnapshotRetentionLimit: snapshotRetentionLimit,
		AppAutoscalingTarget:   targets,
	}

	// Get usage data
	if d.UsageData != nil {
		// Use the helper methods from UsageData to get values
		r.MonthlyDataWrittenGB = d.UsageData.GetFloat("monthly_data_written_gb")
		r.SnapshotStorageSizeGB = d.UsageData.GetFloat("snapshot_storage_size_gb")
		r.ReservedInstanceTerm = d.UsageData.GetString("reserved_instance_term")
		r.ReservedInstancePaymentOption = d.UsageData.GetString("reserved_instance_payment_option")
	}

	return r
}


