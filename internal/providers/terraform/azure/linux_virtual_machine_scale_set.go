package azure

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAzureRMLinuxVirtualMachineScaleSetRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_linux_virtual_machine_scale_set",
		RFunc: NewAzureRMLinuxVirtualMachineScaleSet,
	}
}

func NewAzureRMLinuxVirtualMachineScaleSet(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("location").String()
	instanceType := d.Get("sku").String()

	costComponents := []*schema.CostComponent{linuxVirtualMachineCostComponent(region, instanceType)}
	subResources := make([]*schema.Resource, 0)

	if d.Get("additional_capabilities.0.ultra_ssd_enabled").Bool() {
		costComponents = append(costComponents, ultraSSDReservationCostComponent(region))
	}

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
