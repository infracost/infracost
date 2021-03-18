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
			"Costs associated with non-standard Linux images, such as RHEL are not supported.",
			"Custom machine types are not supported.",
		},
	}
}

func NewAzureRMLinuxVirtualMachine(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {

	region := d.Get("location").String()
	size := d.Get("size").String()
	purchaseOption := "Consumption"

	costComponents := []*schema.CostComponent{
		computeCostComponent(region, size, purchaseOption),
	}

	// todo: add additional cost elements for the vm later on
	// if d.Get("boot_disk.0.initialize_params.0").Exists() {
	// 	costComponents = append(costComponents, bootDisk(region, d.Get("boot_disk.0.initialize_params.0")))
	// }

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func computeCostComponent(region, size string, purchaseOption string) *schema.CostComponent {
	sustainedUseDiscount := 0.0

	sku := strings.ReplaceAll(size, "Standard_", "")
	sku = strings.ReplaceAll(sku, "_", " ")

	return &schema.CostComponent{
		Name:                fmt.Sprintf("Linux/UNIX usage (%s, %s)", purchaseOption, sku),
		Unit:                "hours",
		UnitMultiplier:      1,
		HourlyQuantity:      decimalPtr(decimal.NewFromInt(1)),
		MonthlyDiscountPerc: sustainedUseDiscount,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Virtual Machines"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(sku)},
				{Key: "type", Value: strPtr("Consumption")},
				{Key: "productName", ValueRegex: strPtr(regexDoesNotContain("windows"))},
				{Key: "meterName", ValueRegex: strPtr(regexDoesNotContain("expired"))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
			Unit:           strPtr("1 Hour"),
		},
	}
}
