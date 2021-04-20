package azure

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAzureRMWindowsVirtualMachineScaleSetRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_windows_virtual_machine_scale_set",
		RFunc: NewAzureRMWindowsVirtualMachineScaleSet,
	}
}

func NewAzureRMWindowsVirtualMachineScaleSet(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("location").String()
	instanceType := d.Get("sku").String()
	licenseType := d.Get("license_type").String()

	costComponents := []*schema.CostComponent{windowsVirtualMachineCostComponent(region, instanceType, licenseType)}

	if d.Get("additional_capabilities.0.ultra_ssd_enabled").Bool() {
		costComponents = append(costComponents, ultraSSDReservationCostComponent(region))
	}

	subResources := make([]*schema.Resource, 0)

	osDisk := osDiskSubResource(region, d, u)
	if osDisk != nil {
		subResources = append(subResources, osDisk)
	}

	instanceCount := decimal.NewFromInt(d.Get("instances").Int())
	if u != nil && u.Get("instances").Type != gjson.Null {
		instanceCount = decimal.NewFromInt(u.Get("instances").Int())
	}

	r := &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
		SubResources:   subResources,
	}

	schema.MultiplyQuantities(r, instanceCount)

	return r
}
