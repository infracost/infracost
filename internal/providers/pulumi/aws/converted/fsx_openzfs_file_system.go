package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getFSxOpenZFSFSRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_fsx_openzfs_file_system",
		Notes:     []string{"Data deduplication is not supported by Terraform."},
		RFunc: NewFSxOpenZFSFileSystem,
	}
}
func NewFSxOpenZFSFileSystem(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.FSxOpenZFSFileSystem{
		Address:             d.Address,
		Region:              d.Get("region").String(),
		DeploymentType:      d.Get("deploymentType").String(),
		StorageType:         d.Get("storageType").String(),
		ThroughputCapacity:  d.Get("throughputCapacity").Int(),
		StorageCapacityGB:   d.Get("storageCapacity").Int(),
		ProvisionedIOPS:     d.Get("diskIopsConfiguration.0.iops").Int(),
		ProvisionedIOPSMode: d.Get("diskIopsConfiguration.0.mode").String(),
		DataCompression:     d.Get("rootVolumeConfiguration.0.dataCompressionType").String(),
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
