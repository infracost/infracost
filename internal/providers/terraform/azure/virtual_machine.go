package azure

import (
	"strings"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMVirtualMachineRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_virtual_machine",
		RFunc: NewAzureRMVirtualMachine,
	}
}

func NewAzureRMVirtualMachine(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{})

	costComponents := []*schema.CostComponent{}
	instanceType := d.Get("vm_size").String()

	os := "Linux"
	if d.Get("storage_image_reference.0.offer").Type != gjson.Null {
		if strings.ToLower(d.Get("storage_image_reference.0.offer").String()) == "windowsserver" {
			os = "Windows"
		}
	}
	if strings.ToLower(d.Get("storage_os_disk.0.os_type").String()) == "windows" {
		os = "Windows"
	}

	var monthlyHours *float64 = nil
	if u != nil {
		monthlyHours = u.GetFloat("monthly_hrs")
	}

	if strings.ToLower(os) == "windows" {
		licenseType := d.Get("license_type").String()
		costComponents = append(costComponents, windowsVirtualMachineCostComponent(region, instanceType, licenseType, monthlyHours))
	} else {
		costComponents = append(costComponents, linuxVirtualMachineCostComponent(region, instanceType, monthlyHours))
	}

	// TODO: is this always assuming ultrassdreservation cost?
	costComponents = append(costComponents, ultraSSDReservationCostComponent(region))

	var storageOperations *decimal.Decimal
	if u != nil {
		if v, ok := u.Get("storage_os_disk").Map()["monthly_disk_operations"]; ok {
			storageOperations = decimalPtr(decimal.NewFromInt(v.Int()))
		}
	}

	subResources := []*schema.Resource{}

	if len(d.Get("storage_os_disk").Array()) > 0 {
		diskData := d.Get("storage_os_disk").Array()[0]
		subResources = append(subResources, legacyOSDiskSubResource(region, diskData, storageOperations))
	}

	storages := d.Get("storage_data_disk").Array()
	if u != nil {
		if v, ok := u.Get("storage_data_disk").Map()["monthly_disk_operations"]; ok {
			storageOperations = decimalPtr(decimal.NewFromInt(v.Int()))
		}
	}
	if len(storages) > 0 {
		for _, s := range storages {
			diskType := s.Get("managed_disk_type").String()

			subResources = append(subResources, &schema.Resource{
				Name:           "storage_data_disk",
				CostComponents: managedDiskCostComponents(region, diskType, s, storageOperations),
			})
		}
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
		SubResources:   subResources,
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

func legacyOSDiskSubResource(region string, diskData gjson.Result, monthlyDiskOperations *decimal.Decimal) *schema.Resource {
	diskType := diskData.Get("managed_disk_type").String()

	return &schema.Resource{
		Name:           "storage_os_disk",
		CostComponents: managedDiskCostComponents(region, diskType, diskData, monthlyDiskOperations),
	}
}

func osDiskSubResource(region string, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	if len(d.Get("os_disk").Array()) == 0 {
		return nil
	}

	diskData := d.Get("os_disk").Array()[0]
	diskType := diskData.Get("storage_account_type").String()

	var monthlyDiskOperations *decimal.Decimal

	if u != nil {
		if v, ok := u.Get("os_disk").Map()["monthly_disk_operations"]; ok {
			monthlyDiskOperations = decimalPtr(decimal.NewFromInt(v.Int()))
		}
	}

	return &schema.Resource{
		Name:           "os_disk",
		CostComponents: managedDiskCostComponents(region, diskType, diskData, monthlyDiskOperations),
	}
}
