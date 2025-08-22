package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"

	"github.com/shopspring/decimal"
)

type WindowsVirtualMachine struct {
	Address                               string
	Region                                string
	Size                                  string
	LicenseType                           string
	AdditionalCapabilitiesUltraSSDEnabled bool
	OSDiskData                            *ManagedDiskData
	MonthlyHours                          *float64     `infracost_usage:"monthly_hrs"`
	OSDisk                                *OSDiskUsage `infracost_usage:"os_disk"`
	IsDevTest                             bool
}

type OSDiskUsage struct {
	MonthlyDiskOperations *int64 `infracost_usage:"monthly_disk_operations"`
}

var OSDiskUsageSchema = []*schema.UsageItem{
	{ValueType: schema.Int64, DefaultValue: 0, Key: "monthly_disk_operations"},
}

func (r *WindowsVirtualMachine) CoreType() string {
	return "WindowsVirtualMachine"
}

func (r *WindowsVirtualMachine) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{
			Key:          "os_disk",
			ValueType:    schema.SubResourceUsage,
			DefaultValue: &usage.ResourceUsage{Name: "os_disk", Items: OSDiskUsageSchema},
		},
	}
}

func (r *WindowsVirtualMachine) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *WindowsVirtualMachine) BuildResource() *schema.Resource {
	region := r.Region

	instanceType := r.Size
	licenseType := r.LicenseType

	costComponents := []*schema.CostComponent{windowsVirtualMachineCostComponent(region, instanceType, licenseType, r.MonthlyHours, r.IsDevTest)}

	if r.AdditionalCapabilitiesUltraSSDEnabled {
		costComponents = append(costComponents, ultraSSDReservationCostComponent(region))
	}

	subResources := make([]*schema.Resource, 0)

	var monthlyDiskOperations *decimal.Decimal
	if r.OSDisk != nil && r.OSDisk.MonthlyDiskOperations != nil {
		monthlyDiskOperations = decimalPtr(decimal.NewFromInt(*r.OSDisk.MonthlyDiskOperations))
	}
	if r.OSDiskData != nil {
		osDisk := osDiskSubResource(region, r.OSDiskData.DiskType, r.OSDiskData.DiskSizeGB, r.OSDiskData.DiskIOPSReadWrite, r.OSDiskData.DiskMBPSReadWrite, monthlyDiskOperations)
		if osDisk != nil {
			subResources = append(subResources, osDisk)
		}
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		SubResources:   subResources,
		UsageSchema:    r.UsageSchema(),
	}
}

func windowsVirtualMachineCostComponent(region string, instanceType string, licenseType string, monthlyHours *float64, isDevTest bool) *schema.CostComponent {
	purchaseOption := "Consumption"
	purchaseOptionLabel := "pay as you go"

	productNameRe := "/(Series )?Windows$/i"
	if strings.HasPrefix(instanceType, "Basic_") {
		productNameRe = "/Basic Windows$/"
	} else if !strings.HasPrefix(instanceType, "Standard_") {
		instanceType = fmt.Sprintf("Standard_%s", instanceType)
	}

	if strings.ToLower(licenseType) == "windows_client" || strings.ToLower(licenseType) == "windows_server" {
		purchaseOption = "DevTestConsumption"
		purchaseOptionLabel = "hybrid benefit"
	}

	if isDevTest {
		purchaseOption = "DevTestConsumption"
		purchaseOptionLabel = "dev/test"
	}

	qty := schema.HourToMonthUnitMultiplier
	if monthlyHours != nil {
		qty = decimal.NewFromFloat(*monthlyHours)
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Instance usage (Windows, %s, %s)", purchaseOptionLabel, instanceType),
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
