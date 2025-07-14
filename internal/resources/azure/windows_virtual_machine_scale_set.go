package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"

	"github.com/shopspring/decimal"
)

type WindowsVirtualMachineScaleSet struct {
	Address                               string
	Region                                string
	SKU                                   string
	LicenseType                           string
	AdditionalCapabilitiesUltraSSDEnabled bool
	IsDevTest                             bool
	OSDiskData                            *ManagedDiskData
	Instances                             *int64       `infracost_usage:"instances"`
	OSDisk                                *OSDiskUsage `infracost_usage:"os_disk"`
}

func (r *WindowsVirtualMachineScaleSet) CoreType() string {
	return "WindowsVirtualMachineScaleSet"
}

func (r *WindowsVirtualMachineScaleSet) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "instances", ValueType: schema.Int64, DefaultValue: 0},
		{
			Key:          "os_disk",
			ValueType:    schema.SubResourceUsage,
			DefaultValue: &usage.ResourceUsage{Name: "os_disk", Items: OSDiskUsageSchema},
		},
	}
}

func (r *WindowsVirtualMachineScaleSet) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *WindowsVirtualMachineScaleSet) BuildResource() *schema.Resource {
	region := r.Region

	instanceType := r.SKU
	licenseType := r.LicenseType

	costComponents := []*schema.CostComponent{windowsVirtualMachineCostComponent(region, instanceType, licenseType, nil, r.IsDevTest)}

	if r.AdditionalCapabilitiesUltraSSDEnabled {
		costComponents = append(costComponents, ultraSSDReservationCostComponent(region))
	}

	subResources := make([]*schema.Resource, 0)

	var monthlyDiskOperations *decimal.Decimal
	if r.OSDisk != nil && r.OSDisk.MonthlyDiskOperations != nil {
		monthlyDiskOperations = decimalPtr(decimal.NewFromInt(*r.OSDisk.MonthlyDiskOperations))
	}

	if r.OSDiskData != nil {
		osDisk := osDiskSubResource(region, r.OSDiskData.DiskType, r.OSDiskData.DiskSizeGB, r.OSDiskData.DiskIOPSReadWrite, r.OSDiskData.DiskMBPSReadWrite, monthlyDiskOperations)
		if osDisk != nil {
			subResources = append(subResources, osDisk)
		}
	}

	instanceCount := decimal.NewFromInt(*r.Instances)
	if r.Instances != nil {
		instanceCount = decimal.NewFromInt(*r.Instances)
	}

	res := &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		SubResources:   subResources,
		UsageSchema:    r.UsageSchema(),
	}

	schema.MultiplyQuantities(res, instanceCount)

	return res
}
