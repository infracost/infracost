package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getDMSRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_dms_replication_instance",
		RFunc: NewDmsReplicationInstance,
	}
}
func NewDmsReplicationInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.DmsReplicationInstance{Address: strPtr(d.Address), MultiAz: boolPtr(d.Get("multi_az").Bool()), AllocatedStorage: intPtr(d.Get("allocated_storage").Int()), ReplicationInstanceClass: strPtr(d.Get("replication_instance_class").String()), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
