package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetEFSFileSystemRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "states.aws.efs.file_system.present",
		RFunc: NewEFSFileSystem,
	}
}
func NewEFSFileSystem(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.EFSFileSystem{
		Address: d.Address,
		Region:  d.Get("region").String(),
		//HasLifecyclePolicy:          len(d.Get("lifecycle_policy").Array()) > 0, # lifecycle_policy  is not supported in idem-aws, currently
		AvailabilityZoneName:        d.Get("availability_zone_name").String(),
		ProvisionedThroughputInMBps: d.Get("provisioned_throughput_in_mibps").Float(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
