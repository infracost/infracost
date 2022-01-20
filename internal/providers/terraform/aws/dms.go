package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getDMSRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_dms_replication_instance",
		RFunc: NewDMSReplicationInstance,
	}
}

func NewDMSReplicationInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.DMSReplicationInstance{
		Address:                  d.Address,
		MultiAZ:                  d.Get("multi_az").Bool(),
		AllocatedStorageGB:       d.Get("allocated_storage").Int(),
		ReplicationInstanceClass: d.Get("replication_instance_class").String(),
		Region:                   d.Get("region").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
