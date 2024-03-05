package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getEFSFileSystemRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_efs_file_system",
		CoreRFunc: NewEFSFileSystem,
	}
}
func NewEFSFileSystem(d *schema.ResourceData) schema.CoreResource {
	r := &aws.EFSFileSystem{
		Address:                     d.Address,
		Region:                      d.Get("region").String(),
		HasLifecyclePolicy:          len(d.Get("lifecycle_policy").Array()) > 0,
		AvailabilityZoneName:        d.Get("availability_zone_name").String(),
		ProvisionedThroughputInMBps: d.Get("provisioned_throughput_in_mibps").Float(),
	}
	return r
}
