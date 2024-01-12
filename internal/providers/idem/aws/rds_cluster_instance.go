package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetRDSClusterInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "states.aws.rds.db_instance.present",
		RFunc: NewRDSClusterInstance,
	}
}

func NewRDSClusterInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	piEnabled := d.Get("performance_insights_enabled").Bool()
	piLongTerm := piEnabled && d.Get("performance_insights_retention_period").Int() > 7

	r := &aws.RDSClusterInstance{
		Address:                              d.Address,
		Region:                               d.GetStringOrDefault("region", "us-west-2"),
		InstanceClass:                        d.Get("db_instance_class").String(),
		Engine:                               d.Get("engine").String(),
		PerformanceInsightsEnabled:           piEnabled,
		PerformanceInsightsLongTermRetention: piLongTerm,
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
