package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

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
	instanceType := d.Get("size").String()
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

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
		SubResources:   subResources,
	}
}

func windowsVirtualMachineCostComponent(region string, instanceType string, licenseType string) *schema.CostComponent {
	purchaseOption := "Consumption"
	purchaseOptionLabel := "pay as you go"

	productNameRe := "/Virtual Machines .* Series Windows$/"
	if strings.HasPrefix(instanceType, "Basic_") {
		productNameRe = "/Virtual Machines .* Series Basic Windows$/"
	}

	// Handle Azure Hybrid Benefit
	if licenseType == "Windows_Client" || licenseType == "Windows_Server" {
		purchaseOption = "DevTestConsumption"
		purchaseOptionLabel = "hybrid benefit"
	}

	skuName := parseVMSKUName(instanceType)

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
				{Key: "armSkuName", Value: strPtr(instanceType)},
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
