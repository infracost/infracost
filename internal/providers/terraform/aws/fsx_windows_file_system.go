package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetFSXWindowsFSRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_fsx_windows_file_system",
		Notes: []string{"Data deduplication is not supported by Terraform."},
		RFunc: NewFSXWindowsFS,
	}
}
func NewFSXWindowsFS(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.FSXWindowsFS{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String()), DeploymentType: strPtr(d.Get("deployment_type").String()), StorageType: strPtr(d.Get("storage_type").String()), ThroughputCapacity: intPtr(d.Get("throughput_capacity").Int()), StorageCapacity: intPtr(d.Get("storage_capacity").Int())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
