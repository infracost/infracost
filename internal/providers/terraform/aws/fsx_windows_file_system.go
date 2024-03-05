package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getFSxWindowsFSRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_fsx_windows_file_system",
		Notes:     []string{"Data deduplication is not supported by Terraform."},
		CoreRFunc: NewFSxWindowsFileSystem,
	}
}
func NewFSxWindowsFileSystem(d *schema.ResourceData) schema.CoreResource {
	r := &aws.FSxWindowsFileSystem{
		Address:            d.Address,
		Region:             d.Get("region").String(),
		DeploymentType:     d.Get("deployment_type").String(),
		StorageType:        d.Get("storage_type").String(),
		ThroughputCapacity: d.Get("throughput_capacity").Int(),
		StorageCapacityGB:  d.Get("storage_capacity").Int(),
	}
	return r
}
