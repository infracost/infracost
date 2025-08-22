package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"

	"strings"

	"github.com/shopspring/decimal"
)

type VirtualMachine struct {
	Address                    string
	Region                     string
	StorageImageReferenceOffer string
	VMSize                     string
	StorageOSDiskOSType        string
	LicenseType                string
	StorageOSDiskData          *ManagedDiskData
	OSDiskData                 *ManagedDiskData
	StoragesDiskData           []*ManagedDiskData
	MonthlyHours               *float64              `infracost_usage:"monthly_hrs"`
	StorageOSDisk              *StorageOSDiskUsage   `infracost_usage:"storage_os_disk"`
	StorageDataDisk            *StorageDataDiskUsage `infracost_usage:"storage_data_disk"`
	IsDevTest                  bool
}

type StorageOSDiskUsage struct {
	MonthlyDiskOperations *int64 `infracost_usage:"monthly_disk_operations"`
}

type StorageDataDiskUsage struct {
	MonthlyDiskOperations *int64 `infracost_usage:"monthly_disk_operations"`
}

var StorageOSDiskUsageSchema = []*schema.UsageItem{
	{ValueType: schema.Int64, DefaultValue: 0, Key: "monthly_disk_operations"},
}

var StorageDataDiskUsageSchema = []*schema.UsageItem{
	{ValueType: schema.Int64, DefaultValue: 0, Key: "monthly_disk_operations"},
}

func (r *VirtualMachine) CoreType() string {
	return "VirtualMachine"
}

func (r *VirtualMachine) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{
			Key:          "storage_os_disk",
			ValueType:    schema.SubResourceUsage,
			DefaultValue: &usage.ResourceUsage{Name: "storage_os_disk", Items: StorageOSDiskUsageSchema},
		},
		{
			Key:          "storage_data_disk",
			ValueType:    schema.SubResourceUsage,
			DefaultValue: &usage.ResourceUsage{Name: "storage_data_disk", Items: StorageDataDiskUsageSchema},
		},
	}
}

func (r *VirtualMachine) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *VirtualMachine) BuildResource() *schema.Resource {
	region := r.Region

	costComponents := []*schema.CostComponent{}
	instanceType := r.VMSize

	os := "Linux"
	if r.StorageImageReferenceOffer != "" {
		if strings.ToLower(r.StorageImageReferenceOffer) == "windowsserver" {
			os = "Windows"
		}
	}
	if strings.ToLower(r.StorageOSDiskOSType) == "windows" {
		os = "Windows"
	}

	if strings.ToLower(os) == "windows" {
		licenseType := r.LicenseType
		costComponents = append(costComponents, windowsVirtualMachineCostComponent(region, instanceType, licenseType, r.MonthlyHours, r.IsDevTest))
	} else {
		costComponents = append(costComponents, linuxVirtualMachineCostComponent(region, instanceType, r.MonthlyHours))
	}

	// TODO: is this always assuming ultrassdreservation cost?
	costComponents = append(costComponents, ultraSSDReservationCostComponent(region))

	var storageOperations *decimal.Decimal
	if r.StorageOSDisk != nil && r.StorageOSDisk.MonthlyDiskOperations != nil {
		storageOperations = decimalPtr(decimal.NewFromInt(*r.StorageOSDisk.MonthlyDiskOperations))
	}

	subResources := []*schema.Resource{}

	if r.StorageOSDiskData != nil {
		subResources = append(subResources, legacyOSDiskSubResource(region, r.StorageOSDiskData.DiskType, r.StorageOSDiskData.DiskSizeGB, r.StorageOSDiskData.DiskIOPSReadWrite, r.StorageOSDiskData.DiskMBPSReadWrite, storageOperations))
	}

	if r.StorageOSDisk != nil && r.StorageDataDisk.MonthlyDiskOperations != nil {
		storageOperations = decimalPtr(decimal.NewFromInt(*r.StorageDataDisk.MonthlyDiskOperations))
	}

	for _, s := range r.StoragesDiskData {
		subResources = append(subResources, &schema.Resource{
			Name:           "storage_data_disk",
			CostComponents: managedDiskCostComponents(region, s.DiskType, s.DiskSizeGB, s.DiskIOPSReadWrite, s.DiskMBPSReadWrite, storageOperations),
			UsageSchema:    r.UsageSchema(),
		})
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		SubResources:   subResources,
		UsageSchema:    r.UsageSchema(),
	}
}

func ultraSSDReservationCostComponent(region string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Ultra disk reservation (if unattached)",
		Unit:           "vCPU",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: nil,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Storage"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Ultra Disks")},
				{Key: "skuName", Value: strPtr("Ultra LRS")},
				{Key: "meterName", ValueRegex: regexPtr("Reservation per vCPU Provisioned$")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func legacyOSDiskSubResource(region, diskType string, diskSizeGB, diskIOPSReadWrite, diskMBPSReadWrite int64, monthlyDiskOperations *decimal.Decimal) *schema.Resource {
	return &schema.Resource{
		Name:           "storage_os_disk",
		CostComponents: managedDiskCostComponents(region, diskType, diskSizeGB, diskIOPSReadWrite, diskMBPSReadWrite, monthlyDiskOperations),
	}
}

func osDiskSubResource(region, diskType string, diskSizeGB, diskIOPSReadWrite, diskMBPSReadWrite int64, monthlyDiskOperations *decimal.Decimal) *schema.Resource {
	return &schema.Resource{
		Name:           "os_disk",
		CostComponents: managedDiskCostComponents(region, diskType, diskSizeGB, diskIOPSReadWrite, diskMBPSReadWrite, monthlyDiskOperations),
	}
}
