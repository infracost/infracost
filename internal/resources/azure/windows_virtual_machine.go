package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type WindowsVirtualMachine struct {
	Address                               string
	Region                                string
	Size                                  string
	LicenseType                           string
	AdditionalCapabilitiesUltraSSDEnabled bool
	OSDiskData                            *ManagedDiskData
	MonthlyHours                          *float64     `infracost_usage:"montly_hrs"`
	OSDisk                                *OSDiskUsage `infracost_usage:"os_disk"`
}

type OSDiskUsage struct {
	MonthlyDiskOperations *int64 `infracost_usage:"monthly_disk_operations"`
}

var OSDiskUsageSchema = []*schema.UsageItem{
	{ValueType: schema.Int64, DefaultValue: 0, Key: "monthly_disk_operations"},
}

var WindowsVirtualMachineUsageSchema = []*schema.UsageItem{
	{Key: "montly_hrs", ValueType: schema.Float64, DefaultValue: 0},
	{
		Key:          "os_disk",
		ValueType:    schema.SubResourceUsage,
		DefaultValue: &usage.ResourceUsage{Name: "os_disk", Items: OSDiskUsageSchema},
	},
}

func (r *WindowsVirtualMachine) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *WindowsVirtualMachine) BuildResource() *schema.Resource {
	region := r.Region

	instanceType := r.Size
	licenseType := r.LicenseType

	costComponents := []*schema.CostComponent{windowsVirtualMachineCostComponent(region, instanceType, licenseType, r.MonthlyHours)}

	if r.AdditionalCapabilitiesUltraSSDEnabled {
		costComponents = append(costComponents, ultraSSDReservationCostComponent(region))
	}

	subResources := make([]*schema.Resource, 0)

	var monthlyDiskOperations *decimal.Decimal
	if r.OSDisk.MonthlyDiskOperations != nil {
		monthlyDiskOperations = decimalPtr(decimal.NewFromInt(*r.OSDisk.MonthlyDiskOperations))
	}
	osDisk := osDiskSubResource(region, r.OSDiskData.DiskType, r.OSDiskData.DiskSizeGB, r.OSDiskData.DiskIOPSReadWrite, r.OSDiskData.DiskMBPSReadWrite, monthlyDiskOperations)
	if osDisk != nil {
		subResources = append(subResources, osDisk)
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		SubResources:   subResources, UsageSchema: WindowsVirtualMachineUsageSchema,
	}
}

func windowsVirtualMachineCostComponent(region string, instanceType string, licenseType string, monthlyHours *float64) *schema.CostComponent {
	purchaseOption := "Consumption"
	purchaseOptionLabel := "pay as you go"

	productNameRe := "/Virtual Machines .* Series Windows$/"
	if strings.HasPrefix(instanceType, "Basic_") {
		productNameRe = "/Virtual Machines .* Series Basic Windows$/"
	} else if !strings.HasPrefix(instanceType, "Standard_") {
		instanceType = fmt.Sprintf("Standard_%s", instanceType)
	}

	if strings.ToLower(licenseType) == "windows_client" || strings.ToLower(licenseType) == "windows_server" {
		purchaseOption = "DevTestConsumption"
		purchaseOptionLabel = "hybrid benefit"
	}

	qty := decimal.NewFromFloat(730)
	if monthlyHours != nil {
		qty = decimal.NewFromFloat(*monthlyHours)
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Instance usage (%s, %s)", purchaseOptionLabel, instanceType),
		Unit:            "hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(qty),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Virtual Machines"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", ValueRegex: strPtr("/^(?!.*(Low Priority|Spot)$).*$/i")},
				{Key: "armSkuName", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", instanceType))},
				{Key: "productName", ValueRegex: strPtr(productNameRe)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr(purchaseOption),
			Unit:           strPtr("1 Hour"),
		},
	}
}
