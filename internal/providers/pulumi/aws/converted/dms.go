package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getDMSRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_dms_replication_instance",
		RFunc: NewDMSReplicationInstance,
	}
}

func NewDMSReplicationInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.DMSReplicationInstance{
		Address:                  d.Address,
		MultiAZ:                  d.Get("multiAz").Bool(),
		AllocatedStorageGB:       d.Get("allocatedStorage").Int(),
		ReplicationInstanceClass: d.Get("replicationInstanceClass").String(),
		Region:                   d.Get("region").String(),
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
