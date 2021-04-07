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
			"Low priority, Spot and Reserved instances are not supported.",
		},
	}
}

func NewAzureRMWindowsVirtualMachine(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("location").String()

	costComponents := []*schema.CostComponent{windowsVirtualMachineCostComponent(d)}
	subResources := make([]*schema.Resource, 0)
	if len(d.Get("os_disk").Array()) > 0 {
		subResources = append(subResources, osDiskSubResource(region, d.Get("os_disk").Array()[0], u))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
		SubResources:   subResources,
	}
}

func windowsVirtualMachineCostComponent(d *schema.ResourceData) *schema.CostComponent {
	region := d.Get("location").String()
	size := d.Get("size").String()
	purchaseOption := "Consumption"
	purchaseOptionLabel := "pay as you go"

	productNameRe := "/Virtual Machines .* Series Windows$/"
	if strings.HasPrefix(size, "Basic_") {
		productNameRe = "/Virtual Machines .* Series Basic Windows$/"
	}

	// Handle Azure Hybrid Benefit
	licenseType := d.Get("license_type").String()
	if licenseType == "Windows_Client" || licenseType == "Windows_Server" {
		purchaseOption = "DevTestConsumption"
		purchaseOptionLabel = "hybrid benefit"
	}

	skuName := parseVMSKUName(size)

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Instance usage (%s, %s)", purchaseOptionLabel, skuName),
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
				{Key: "skuName", Value: strPtr(skuName)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr(purchaseOption),
			Unit:           strPtr("1 Hour"),
		},
	}
}
