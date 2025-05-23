package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getRDSClusterInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_rds_cluster_instance",
		RFunc: NewRDSClusterInstance,
	}
}

func NewRDSClusterInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	piEnabled := d.Get("performanceInsightsEnabled").Bool()
	piLongTerm := piEnabled && d.Get("performanceInsightsRetentionPeriod").Int() > 7

	r := &aws.RDSClusterInstance{
		Address:                              d.Address,
		Region:                               d.Get("region").String(),
		InstanceClass:                        d.Get("instanceClass").String(),
		Engine:                               d.Get("engine").String(),
		PerformanceInsightsEnabled:           piEnabled,
		PerformanceInsightsLongTermRetention: piLongTerm,
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
