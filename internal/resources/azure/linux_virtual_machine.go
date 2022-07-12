package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type LinuxVirtualMachine struct {
	Address         string
	Region          string
	Size            string
	UltraSSDEnabled bool
	MonthlyHrs      *float64 `infracost_usage:"monthly_hrs"`
}

var LinuxVirtualMachineUsageSchema = []*schema.UsageItem{{Key: "monthly_hrs", ValueType: schema.Float64, DefaultValue: 0}}

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

	osDisk := osDiskSubResource(r.Region)
	if osDisk != nil {
		subResources = append(subResources, osDisk)
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		SubResources:   subResources, UsageSchema: LinuxVirtualMachineUsageSchema,
	}
}

func linuxVirtualMachineCostComponent(region string, instanceType string, monthlyHours *float64) *schema.CostComponent {
	purchaseOption := "Consumption"
	purchaseOptionLabel := "pay as you go"

	productNameRe := "/Virtual Machines .* Series$/"
	if strings.HasPrefix(instanceType, "Basic_") {
		productNameRe = "/Virtual Machines .* Series Basic$/"
	} else if !strings.HasPrefix(instanceType, "Standard_") {
		instanceType = fmt.Sprintf("Standard_%s", instanceType)
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
				{Key: "meterName", ValueRegex: strPtr("/^(?!.*(Expired|Free)$).*$/i")},
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
