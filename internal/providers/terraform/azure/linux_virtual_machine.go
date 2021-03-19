package azure

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

// todo: add devtest consumption
func GetAzureRMLinuxVirtualMachineRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_linux_virtual_machine",
		RFunc: NewAzureRMLinuxVirtualMachine,
		Notes: []string{
			"Costs associated with standard Linux images.",
			"Non-Standard images such as RHEL are not supported.",
			"Only Standard machine types are currently supported.",
			"Only Pay-As-You-Go consumption prices are currently supported.",
		},
	}
}

func NewAzureRMLinuxVirtualMachine(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: linuxVirtualMachineCostComponent(d),
	}
}

func linuxVirtualMachineCostComponent(d *schema.ResourceData) []*schema.CostComponent {
	purchaseOption := "Consumption"
	region := d.Get("location").String()
	size := d.Get("size").String()

	// todo: add additional cost elements for the vm later on
	// if d.Get("boot_disk.0.initialize_params.0").Exists() {
	// 	costComponents = append(costComponents, bootDisk(region, d.Get("boot_disk.0.initialize_params.0")))
	// }

	sku := parseVirtualMachineSizeSKU(size)

	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, &schema.CostComponent{
		Name:           fmt.Sprintf("(%s, %s)", purchaseOption, sku),
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Virtual Machines"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(sku)},
				{Key: "type", Value: strPtr("Consumption")},
				{Key: "productName", ValueRegex: strPtr(regexMustNotContain("windows"))},
				{Key: "meterName", ValueRegex: strPtr(regexMustNotContain("expired"))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
			Unit:           strPtr("1 Hour"),
		},
	})

	return costComponents
}
