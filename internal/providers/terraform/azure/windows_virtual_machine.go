package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

// todo: add devtest consumption
func GetAzureRMWindowsVirtualMachineRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_windows_virtual_machine",
		RFunc: NewAzureRMWindowsVirtualMachine,
		Notes: []string{
			"Costs associated with Windows Server virtual machines.",
			"Only Standard machine types are currently supported.",
			"Only Pay-As-You-Go consumption prices are currently supported (No BYOL).",
		},
	}
}

func NewAzureRMWindowsVirtualMachine(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: []*schema.CostComponent{windowsVirtualMachineCostComponent(d)},
	}
}

func windowsVirtualMachineCostComponent(d *schema.ResourceData) *schema.CostComponent {
	region := d.Get("location").String()
	size := d.Get("size").String()
	purchaseOption := "Consumption"
	purchaseOptionLabel := "pay as you go"

	productNameRe := "/Virtual Machines .* Series Windows$/"
	if strings.HasPrefix(size, "Basic_") {
		productNameRe = "/Virtual Machines .* Series Basic Windows $/"
	}

	// Handle Azure Hybrid Benefit
	licenseType := d.Get("license_type").String()
	if licenseType == "Windows_Client" || licenseType == "Windows_Server" {
		purchaseOption = "DevTestConsumption"
		purchaseOptionLabel = "hybrid benefit"
	}

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Instance usage (%s, %s)", purchaseOptionLabel, parseInstanceType(size)),
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Virtual Machines"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "armSkuName", Value: strPtr(size)},
				{Key: "productName", ValueRegex: strPtr(productNameRe)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr(purchaseOption),
			Unit:           strPtr("1 Hour"),
		},
	}
}
