package azure

import (
	"strings"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMVirtualMachineScaleSetRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_virtual_machine_scale_set",
		RFunc: NewAzureRMVirtualMachineScaleSet,
	}
}

func NewAzureRMVirtualMachineScaleSet(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{})

	costComponents := []*schema.CostComponent{}
	subResources := []*schema.Resource{}

	instanceType := d.Get("sku.0.name").String()
	capacity := decimal.NewFromInt(d.Get("sku.0.capacity").Int())

	if u != nil && u.Get("instances").Type != gjson.Null {
		capacity = decimal.NewFromInt(u.Get("instances").Int())
	}

	os := "Linux"
	if !d.IsEmpty("os_profile_windows_config") {
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
		costComponents = append(costComponents, linuxVirtualMachineCostComponent(region, instanceType, nil))
	}

	if strings.ToLower(os) == "windows" {
		licenseType := "Windows_Client"
		if d.Get("license_type").Type != gjson.Null {
			licenseType = d.Get("license_type").String()
		}
		costComponents = append(costComponents, windowsVirtualMachineCostComponent(region, instanceType, licenseType, nil))
	}

	r := &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
		SubResources:   subResources,
	}

	schema.MultiplyQuantities(r, capacity)

	var storageOperations *decimal.Decimal
	if len(d.Get("storage_profile_os_disk").Array()) > 0 {
		diskData := d.Get("storage_profile_os_disk").Array()[0]
		if u != nil {
			if v, ok := u.Get("storage_profile_os_disk").Map()["monthly_disk_operations"]; ok {
				storageOperations = decimalPtr(decimal.NewFromInt(v.Int()))
			}
		}
		r.SubResources = append(r.SubResources, legacyOSDiskSubResource(region, diskData, storageOperations))
	}

	if u != nil {
		if v, ok := u.Get("storage_profile_data_disk").Map()["monthly_disk_operations"]; ok {
			storageOperations = decimalPtr(decimal.NewFromInt(v.Int()))
		}
	}

	storages := d.Get("storage_profile_data_disk").Array()
	if len(storages) > 0 {
		for _, s := range storages {
			if s.Get("managed_disk_type").Type != gjson.Null {
				diskType := s.Get("managed_disk_type").String()

				r.SubResources = append(r.SubResources, &schema.Resource{
					Name:           "storage_data_disk",
					CostComponents: managedDiskCostComponents(region, diskType, s, storageOperations),
				})
			}
		}
	}

	return r
}
