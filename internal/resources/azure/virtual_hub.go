package azure

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// VirtualHub is the central hub in the "hub and spoke architecture" of Azure Virtual WAN.
// It enables transitive connectivity between endpoints that may be distributed across different types of 'spokes'.
//
// More resource information here: https://docs.microsoft.com/en-us/azure/virtual-wan/virtual-wan-about
// Pricing information here: https://azure.microsoft.com/en-us/pricing/details/virtual-wan/
type VirtualHub struct {
	// Address is the unique name of the resource in the IAC language.
	Address string
	// Region is the azure region the VirtualHub is provisioned within.
	Region string
	// SKU is the VirtualHub hub type. It can be one of: Basic|Standard.
	SKU string

	// MonthlyDataProcessedGB represents a usage cost for the amount of gb of data that is processed
	// through the hub on a monthly basis. It is a float to allow users to specify values whole GBs.
	MonthlyDataProcessedGB *float64 `infracost_usage:"monthly_data_processed_gb"`
}

// CoreType returns the name of this resource type
func (v *VirtualHub) CoreType() string {
	return "VirtualHub"
}

// UsageSchema defines a list which represents the usage schema of VirtualHubUsageSchema.
func (v *VirtualHub) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_data_processed_gb", DefaultValue: 0, ValueType: schema.Float64},
	}
}

// PopulateUsage parses the u schema.UsageData into the VirtualHub.
// It uses the `infracost_usage` struct tags to populate data into the VirtualHub.
func (v *VirtualHub) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(v, u)
}

// BuildResource builds a schema.Resource from a valid VirtualHub.
// It returns VirtualHub as a *schema.Resource with 2 cost components provided.
// These cost components are only applicable if the VirtualHub is type Standard.
// The Basic hub is provided free by azure.
// See here: https://azure.microsoft.com/en-us/pricing/details/virtual-wan/ for more information.
//
// This method is called after the resource is initialised by an iac provider.
// See providers folder for more information.
func (v *VirtualHub) BuildResource() *schema.Resource {
	if v.SKU == "Basic" {
		return &schema.Resource{
			Name:        v.Address,
			UsageSchema: v.UsageSchema(),
			NoPrice:     true,
			IsSkipped:   true,
		}
	}

	components := []*schema.CostComponent{
		v.deploymentHours(),
	}

	if v.MonthlyDataProcessedGB != nil {
		components = append(components, v.dataProcessed())
	}

	return &schema.Resource{
		Name:           v.Address,
		UsageSchema:    v.UsageSchema(),
		CostComponents: components,
	}
}

func (v VirtualHub) dataProcessed() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Data processed",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromFloat(*v.MonthlyDataProcessedGB)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(v.Region),
			Service:       strPtr("Virtual WAN"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: v.hubTypeFilter()},
				{Key: "meterName", Value: v.hubTypeDataFilter()},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
		UsageBased: true,
	}
}

func (v VirtualHub) deploymentHours() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Deployment",
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(v.Region),
			Service:       strPtr("Virtual WAN"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: v.hubTypeFilter()},
				{Key: "meterName", Value: v.hubTypeUnitFilter()},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func (v VirtualHub) hubTypeFilter() *string {
	return strPtr(fmt.Sprintf("%s Hub", v.SKU))
}

func (v VirtualHub) hubTypeUnitFilter() *string {
	return strPtr(fmt.Sprintf("%s Hub Unit", v.SKU))
}

func (v VirtualHub) hubTypeDataFilter() *string {
	return strPtr(fmt.Sprintf("%s Hub Data Processed", v.SKU))
}
