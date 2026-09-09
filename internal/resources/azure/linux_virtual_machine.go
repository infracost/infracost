package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"

	"github.com/shopspring/decimal"
)

type LinuxVirtualMachine struct {
	Address         string
	Region          string
	Size            string
	UltraSSDEnabled bool
	OSDiskData      *ManagedDiskData
	OSDisk          *OSDiskUsage `infracost_usage:"os_disk"`
	MonthlyHrs      *float64     `infracost_usage:"monthly_hrs"`
}

func (r *LinuxVirtualMachine) CoreType() string {
	return "LinuxVirtualMachine"
}

func (r *LinuxVirtualMachine) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_hrs", ValueType: schema.Float64, DefaultValue: 0},
		{
			Key:          "os_disk",
			ValueType:    schema.SubResourceUsage,
			DefaultValue: &usage.ResourceUsage{Name: "os_disk", Items: OSDiskUsageSchema},
		},
	}
}

func (r *LinuxVirtualMachine) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *LinuxVirtualMachine) BuildResource() *schema.Resource {
	instanceType := r.Size

	costComponents := []*schema.CostComponent{linuxVirtualMachineCostComponent(r.Region, instanceType, r.MonthlyHrs)}

	if r.UltraSSDEnabled {
		costComponents = append(costComponents, ultraSSDReservationCostComponent(r.Region))
	}

	subResources := make([]*schema.Resource, 0)

	var monthlyDiskOperations *decimal.Decimal
	if r.OSDisk != nil && r.OSDisk.MonthlyDiskOperations != nil {
		monthlyDiskOperations = decimalPtr(decimal.NewFromInt(*r.OSDisk.MonthlyDiskOperations))
	}

	if r.OSDiskData != nil {
		osDisk := osDiskSubResource(r.Region, r.OSDiskData.DiskType, r.OSDiskData.DiskSizeGB, r.OSDiskData.DiskIOPSReadWrite, r.OSDiskData.DiskMBPSReadWrite, monthlyDiskOperations)
		subResources = append(subResources, osDisk)
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		SubResources:   subResources,
		UsageSchema:    r.UsageSchema(),
	}
}

func linuxVirtualMachineCostComponent(region string, instanceType string, monthlyHours *float64) *schema.CostComponent {
	return linuxVirtualMachineCostComponentWithPriority(region, instanceType, monthlyHours, "")
}

// linuxVirtualMachineCostComponentWithPriority returns the per-hour instance
// cost component for a Linux VM/node, optionally selecting Spot pricing.
//
// priority="Spot" matches SKUs ending with " Spot" in the Azure retail catalog
// (variable market price; the catalog snapshot is used as a point estimate
// — actual Spot prices fluctuate with demand).
// priority="" or "Regular" excludes Low Priority/Spot SKUs, falling back to
// the standard pay-as-you-go price.
func linuxVirtualMachineCostComponentWithPriority(region string, instanceType string, monthlyHours *float64, priority string) *schema.CostComponent {
	purchaseOption := "Consumption"
	purchaseOptionLabel := "pay as you go"
	skuNameRe := "/^(?!.*(Low Priority|Spot)$).*$/i"
	if strings.EqualFold(priority, "Spot") {
		purchaseOption = "Spot"
		purchaseOptionLabel = "spot"
		skuNameRe = "/^.*Spot$/i"
	}

	productNameRe := "/Series( Linux)?$/i"
	if strings.HasPrefix(strings.ToLower(instanceType), "basic_") {
		productNameRe = "/Series Basic$/"
	} else if !strings.HasPrefix(strings.ToLower(instanceType), "standard_") {
		instanceType = fmt.Sprintf("Standard_%s", instanceType)
	}

	qty := schema.HourToMonthUnitMultiplier
	if monthlyHours != nil {
		qty = decimal.NewFromFloat(*monthlyHours)
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Instance usage (Linux, %s, %s)", purchaseOptionLabel, instanceType),
		Unit:            "hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(qty),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Virtual Machines"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", ValueRegex: strPtr("/^(?!.*(Expired|Free)$).*$/i")},
				{Key: "skuName", ValueRegex: strPtr(skuNameRe)},
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
