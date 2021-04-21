package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetAzureRMLinuxVirtualMachineRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_linux_virtual_machine",
		RFunc: NewAzureRMLinuxVirtualMachine,
		Notes: []string{
			"Non-standard images such as RHEL are not supported.",
			"Low priority, Spot and Reserved instances are not supported.",
		},
	}
}

func NewAzureRMLinuxVirtualMachine(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("location").String()
	instanceType := d.Get("size").String()

	costComponents := []*schema.CostComponent{linuxVirtualMachineCostComponent(region, instanceType)}

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

func linuxVirtualMachineCostComponent(region string, instanceType string) *schema.CostComponent {
	purchaseOption := "Consumption"
	purchaseOptionLabel := "pay as you go"

	productNameRe := "/Virtual Machines .* Series$/"
	if strings.HasPrefix(instanceType, "Basic_") {
		productNameRe = "/Virtual Machines .* Series Basic$/"
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
