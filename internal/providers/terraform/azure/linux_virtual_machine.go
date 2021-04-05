package azure

import (
	"fmt"
	"strings"

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
		CostComponents: []*schema.CostComponent{linuxVirtualMachineCostComponent(d)},
	}
}

func linuxVirtualMachineCostComponent(d *schema.ResourceData) *schema.CostComponent {
	region := d.Get("location").String()
	size := d.Get("size").String()
	purchaseOption := "Consumption"
	purchaseOptionLabel := "pay as you go"

	productNameRe := "/Virtual Machines .* Series$/"
	if strings.HasPrefix(size, "Basic_") {
		productNameRe = "/Virtual Machines .* Series Basic$/"
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
