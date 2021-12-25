package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetDMSRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_dms_replication_instance",
		RFunc: NewDMS,
	}
}
func NewDMS(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.DMS{Address: strPtr(d.Address), ReplicationInstanceClass: strPtr(d.Get("replication_instance_class").String()), Region: strPtr(d.Get("region").String()), MultiAz: boolPtr(d.Get("multi_az").Bool()), AllocatedStorage: intPtr(d.Get("allocated_storage").Int())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
