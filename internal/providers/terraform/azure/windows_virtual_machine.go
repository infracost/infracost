package azure

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

// todo: add devtest consumption
func GetAzureRMWindowsVirtualMachineRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_windows_virtual_machine",
		RFunc: NewAzureRMWindowsVirtualMachine,
		Notes: []string{
			"Costs associated with non-standard Windows images such as RHEL are not supported.",
			"Only Standard machine types are not supported.",
			"Only Pay-As-You-Go consumption prices are currently supported.",
		},
	}
}

func NewAzureRMWindowsVirtualMachine(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: windowsVirtualMachineCostComponent(d),
	}
}

func windowsVirtualMachineCostComponent(d *schema.ResourceData) []*schema.CostComponent {
	purchaseOption := "Consumption"
	region := d.Get("location").String()
	size := d.Get("size").String()

	// todo: add additional cost elements for the vm later on

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
				{Key: "type", Value: strPtr(purchaseOption)},
				{Key: "productName", ValueRegex: strPtr(regexMustContain("windows"))},
				{Key: "meterName", ValueRegex: strPtr(regexMustNotContain("expired"))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr(purchaseOption),
			Unit:           strPtr("1 Hour"),
		},
	})

	return costComponents
}
