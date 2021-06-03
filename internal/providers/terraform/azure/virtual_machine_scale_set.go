package azure

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"strings"
)

func GetAzureRMVirtualMachineScaleSetRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_virtual_machine_scale_set",
		RFunc: NewAzureRMVirtualMachineScaleSet,
	}
}

func NewAzureRMVirtualMachineScaleSet(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	costComponents := []*schema.CostComponent{}
	subResources := []*schema.Resource{}

	location := d.Get("location").String()
	instanceType := d.Get("sku.0.name").String()
	capacity := decimal.NewFromInt(d.Get("sku.0.capacity").Int())

	if u != nil && u.Get("instances").Type != gjson.Null {
		capacity = decimal.NewFromInt(u.Get("instances").Int())
	}

	os := "Linux"
	if d.Get("os_profile_windows_config").Type != gjson.Null {
		os = "Windows"
	}
	if d.Get("storage_profile_os_disk.0.os_type").Type != gjson.Null {
		if strings.ToLower(d.Get("storage_profile_os_disk.0.os_type").String()) == "windows" {
			os = "Windows"
		}
	}
	if d.Get("storage_profile_image_reference.0.offer").Type != gjson.Null {
		if strings.ToLower(d.Get("storage_profile_image_reference.0.offer").String()) == "windowsserver" {
			os = "Windows"
		}
	}

	if strings.ToLower(os) == "linux" {
		costComponents = append(costComponents, linuxVirtualMachineCostComponent(location, instanceType))
	}

	if strings.ToLower(os) == "windows" {
		licenseType := "Windows_Client"
		if d.Get("license_type").Type != gjson.Null {
			licenseType = d.Get("license_type").String()
		}
		costComponents = append(costComponents, windowsVirtualMachineCostComponent(location, instanceType, licenseType))
	}

	r := &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
		SubResources:   subResources,
	}

	schema.MultiplyQuantities(r, capacity)

	diskData := d.Get("storage_profile_os_disk").Array()[0]
	var storageOperations *decimal.Decimal
	if u != nil && u.Get("storage_profile_os_disk.monthly_disk_operations").Type != gjson.Null {
		storageOperations = decimalPtr(decimal.NewFromInt(u.Get("storage_profile_os_disk.monthly_disk_operations").Int()))
	}
	r.SubResources = append(r.SubResources, legacyOSDiskSubResource(location, diskData, storageOperations))

	if u != nil && u.Get("storage_profile_data_disk.monthly_disk_operations").Type != gjson.Null {
		storageOperations = decimalPtr(decimal.NewFromInt(u.Get("storage_profile_data_disk.monthly_disk_operations").Int()))
	}

	storages := d.Get("storage_profile_data_disk").Array()
	if u != nil && u.Get("storage_profile_data_disk.monthly_disk_operations").Type != gjson.Null {
		storageOperations = decimalPtr(decimal.NewFromInt(u.Get("storage_profile_data_disk.monthly_disk_operations").Int()))
	}
	if len(storages) > 0 {
		for _, s := range storages {
			if s.Get("managed_disk_type").Type != gjson.Null {
				diskType := s.Get("managed_disk_type").String()

				r.SubResources = append(r.SubResources, &schema.Resource{
					Name:           "storage_data_disk",
					CostComponents: managedDiskCostComponents(location, diskType, s, storageOperations),
				})
			}
		}
	}

	return r
}
