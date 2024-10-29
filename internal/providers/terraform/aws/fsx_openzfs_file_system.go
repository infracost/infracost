package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getFSxOpenZFSFSRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_fsx_openzfs_file_system",
		Notes:     []string{"Data deduplication is not supported by Terraform."},
		CoreRFunc: NewFSxOpenZFSFileSystem,
	}
}
func NewFSxOpenZFSFileSystem(d *schema.ResourceData) schema.CoreResource {
	r := &aws.FSxOpenZFSFileSystem{
		Address:             d.Address,
		Region:              d.Get("region").String(),
		DeploymentType:      d.Get("deployment_type").String(),
		StorageType:         d.Get("storage_type").String(),
		ThroughputCapacity:  d.Get("throughput_capacity").Int(),
		StorageCapacityGB:   d.Get("storage_capacity").Int(),
		ProvisionedIOPS:     d.Get("disk_iops_configuration.0.iops").Int(),
		ProvisionedIOPSMode: d.Get("disk_iops_configuration.0.mode").String(),
		DataCompression:     d.Get("root_volume_configuration.0.data_compression_type").String(),
	}
	return r
}
