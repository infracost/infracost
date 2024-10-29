package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getRDSClusterInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_rds_cluster_instance",
		CoreRFunc:           NewRDSClusterInstance,
		ReferenceAttributes: []string{"cluster_identifier"},
	}
}

func NewRDSClusterInstance(d *schema.ResourceData) schema.CoreResource {
	piEnabled := d.Get("performance_insights_enabled").Bool()
	piLongTerm := piEnabled && d.Get("performance_insights_retention_period").Int() > 7

	ioOptimized := false
	clusterRefs := d.References("cluster_identifier")
	if len(clusterRefs) > 0 {
		ioOptimized = clusterRefs[0].Get("storage_type").String() == "aurora-iopt1"
	}

	r := &aws.RDSClusterInstance{
		Address:                              d.Address,
		Region:                               d.Get("region").String(),
		InstanceClass:                        d.Get("instance_class").String(),
		IOOptimized:                          ioOptimized,
		Engine:                               d.Get("engine").String(),
		Version:                              d.Get("engine_version").String(),
		PerformanceInsightsEnabled:           piEnabled,
		PerformanceInsightsLongTermRetention: piLongTerm,
	}
	return r
}
