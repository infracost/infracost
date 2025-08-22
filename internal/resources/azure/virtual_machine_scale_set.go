package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"

	"strings"

	"github.com/shopspring/decimal"
)

type VirtualMachineScaleSet struct {
	Address                   string
	Region                    string
	SKUName                   string
	SKUCapacity               int64
	IsWindows                 bool
	IsDevTest                 bool
	LicenseType               string
	StorageProfileOSDiskData  *ManagedDiskData
	StorageProfileOSDisksData []*ManagedDiskData

	Instances              *int64                     `infracost_usage:"instances"`
	StorageProfileOSDisk   *StorageProfileOSDiskUsage `infracost_usage:"storage_profile_os_disk"`
	StorageProfileDataDisk *StorageProfileOSDiskUsage `infracost_usage:"storage_profile_data_disk"`
}

type StorageProfileOSDiskUsage struct {
	MonthlyDiskOperations *int64 `infracost_usage:"monthly_disk_operations"`
}

type StorageProfileDataDiskUsage struct {
	MonthlyDiskOperations *int64 `infracost_usage:"monthly_disk_operations"`
}

var StorageProfileOSDiskUsageSchema = []*schema.UsageItem{
	{ValueType: schema.Int64, DefaultValue: 0, Key: "monthly_disk_operations"},
}

var StorageProfileDataDiskUsageSchema = []*schema.UsageItem{
	{ValueType: schema.Int64, DefaultValue: 0, Key: "monthly_disk_operations"},
}

func (r *VirtualMachineScaleSet) CoreType() string {
	return "VirtualMachineScaleSet"
}

func (r *VirtualMachineScaleSet) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "instances", ValueType: schema.Int64, DefaultValue: 0},
		{
			Key:          "storage_profile_os_disk",
			ValueType:    schema.SubResourceUsage,
			DefaultValue: &usage.ResourceUsage{Name: "storage_profile_os_disk", Items: StorageProfileOSDiskUsageSchema},
		},
		{
			Key:          "storage_profile_data_disk",
			ValueType:    schema.SubResourceUsage,
			DefaultValue: &usage.ResourceUsage{Name: "storage_profile_data_disk", Items: StorageProfileDataDiskUsageSchema},
		},
	}
}

func (r *VirtualMachineScaleSet) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *VirtualMachineScaleSet) BuildResource() *schema.Resource {
	region := r.Region

	costComponents := []*schema.CostComponent{}
	subResources := []*schema.Resource{}

	instanceType := r.SKUName
	capacity := decimal.NewFromInt(r.SKUCapacity)

	if r.Instances != nil {
		capacity = decimal.NewFromInt(*r.Instances)
	}

	os := "Linux"
	if r.IsWindows {
		os = "Windows"
	}

	if strings.ToLower(os) == "linux" {
		costComponents = append(costComponents, linuxVirtualMachineCostComponent(region, instanceType, nil))
	}

	if strings.ToLower(os) == "windows" {
		licenseType := "Windows_Client"
		if r.LicenseType != "" {
			licenseType = r.LicenseType
		}
		costComponents = append(costComponents, windowsVirtualMachineCostComponent(region, instanceType, licenseType, nil, r.IsDevTest))
	}

	res := &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		SubResources:   subResources,
		UsageSchema:    r.UsageSchema(),
	}

	schema.MultiplyQuantities(res, capacity)

	var storageOperations *decimal.Decimal
	if r.StorageProfileOSDisk != nil && r.StorageProfileOSDisk.MonthlyDiskOperations != nil {
		storageOperations = decimalPtr(decimal.NewFromInt(*r.StorageProfileOSDisk.MonthlyDiskOperations))
	}
	if r.StorageProfileOSDiskData != nil {
		res.SubResources = append(res.SubResources, legacyOSDiskSubResource(region, r.StorageProfileOSDiskData.DiskType, r.StorageProfileOSDiskData.DiskSizeGB, r.StorageProfileOSDiskData.DiskIOPSReadWrite, r.StorageProfileOSDiskData.DiskMBPSReadWrite, storageOperations))
	}

	if r.StorageProfileDataDisk != nil && r.StorageProfileDataDisk.MonthlyDiskOperations != nil {
		storageOperations = decimalPtr(decimal.NewFromInt(*r.StorageProfileDataDisk.MonthlyDiskOperations))
	}

	for _, s := range r.StorageProfileOSDisksData {
		res.SubResources = append(res.SubResources, &schema.Resource{
			Name:           "storage_data_disk",
			CostComponents: managedDiskCostComponents(region, s.DiskType, s.DiskSizeGB, s.DiskIOPSReadWrite, s.DiskMBPSReadWrite, storageOperations),
			UsageSchema:    r.UsageSchema(),
		})
	}

	return res
}
