package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getFSxWindowsFSRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_fsx_windows_file_system",
		Notes:     []string{"Data deduplication is not supported by Terraform."},
		RFunc: NewFSxWindowsFileSystem,
	}
}
func NewFSxWindowsFileSystem(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.FSxWindowsFileSystem{
		Address:            d.Address,
		Region:             d.Get("region").String(),
		DeploymentType:     d.Get("deploymentType").String(),
		StorageType:        d.Get("storageType").String(),
		ThroughputCapacity: d.Get("throughputCapacity").Int(),
		StorageCapacityGB:  d.Get("storageCapacity").Int(),
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
