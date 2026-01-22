package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// ContainerAppEnvironment struct represents an Azure Container Apps Environment.
//
// Resource information: https://learn.microsoft.com/en-us/azure/container-apps/environment
// Pricing information: https://azure.microsoft.com/en-us/pricing/details/container-apps/
type ContainerAppEnvironment struct {
	Address          string
	Region           string
	WorkloadProfiles []ContainerAppEnvironmentWorkloadProfile
}

type ContainerAppEnvironmentWorkloadProfile struct {
	Name                string
	WorkloadProfileType string
	MinimumCount        int64
	MaximumCount        int64
}

var (
	containerAppProfileSpecs = map[string]struct {
		vCPU   float64
		memory float64
	}{
		"D4":  {vCPU: 4, memory: 16},
		"D8":  {vCPU: 8, memory: 32},
		"D16": {vCPU: 16, memory: 64},
		"D32": {vCPU: 32, memory: 128},
		"E4":  {vCPU: 4, memory: 32},
		"E8":  {vCPU: 8, memory: 64},
		"E16": {vCPU: 16, memory: 128},
		"E32": {vCPU: 32, memory: 256},
	}
)

func (r *ContainerAppEnvironment) CoreType() string {
	return "ContainerAppEnvironment"
}

func (r *ContainerAppEnvironment) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *ContainerAppEnvironment) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ContainerAppEnvironment) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	// Check if we have any dedicated workload profiles
	hasDedicatedProfile := false

	for _, profile := range r.WorkloadProfiles {
		if strings.EqualFold(profile.WorkloadProfileType, "Consumption") {
			continue
		}

		hasDedicatedProfile = true

		spec, ok := containerAppProfileSpecs[strings.ToUpper(profile.WorkloadProfileType)]
		if !ok {
			// Skip unknown profiles
			continue
		}

		if profile.MinimumCount > 0 {
			costComponents = append(costComponents, r.dedicatedVCPUCostComponent(profile, spec.vCPU))
			costComponents = append(costComponents, r.dedicatedMemoryCostComponent(profile, spec.memory))
		}
	}

	if hasDedicatedProfile {
		costComponents = append(costComponents, r.managementFeeCostComponent())
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *ContainerAppEnvironment) dedicatedVCPUCostComponent(profile ContainerAppEnvironmentWorkloadProfile, vCPU float64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Dedicated vCPU usage (%s)", profile.WorkloadProfileType),
		Unit:            "hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(profile.MinimumCount).Mul(decimal.NewFromFloat(vCPU)).Mul(decimal.NewFromInt(730))),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Azure Container Apps"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Azure Container Apps")},
				{Key: "skuName", Value: strPtr("Dedicated")},
				{Key: "meterName", Value: strPtr("Dedicated vCPU Usage")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func (r *ContainerAppEnvironment) dedicatedMemoryCostComponent(profile ContainerAppEnvironmentWorkloadProfile, memory float64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Dedicated memory usage (%s)", profile.WorkloadProfileType),
		Unit:            "GiB-hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(profile.MinimumCount).Mul(decimal.NewFromFloat(memory)).Mul(decimal.NewFromInt(730))),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Azure Container Apps"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Azure Container Apps")},
				{Key: "skuName", Value: strPtr("Dedicated")},
				{Key: "meterName", Value: strPtr("Dedicated Memory Usage")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func (r *ContainerAppEnvironment) managementFeeCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Management fee",
		Unit:            "hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(730)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Azure Container Apps"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr("Azure Container Apps")},
				{Key: "skuName", Value: strPtr("Dedicated")},
				{Key: "meterName", Value: strPtr("Dedicated Plan Management")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
