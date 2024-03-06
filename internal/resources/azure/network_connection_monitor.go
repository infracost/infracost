package azure

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
)

// NetworkConnectionMonitor struct represents Azure Network Watcher Connection
// Monitor (new)
//
// This resource is charged for each test. Connection Monitors can have
// multiple test groups. A test is a combination of a source endpoint,
// destination endpoint and test configuration within a test group. The number
// of tests is calculated by multiplying the number of source endpoints,
// destination endpoints and test configurations for each enabled test group.
//
// There is a free limit of 1000 tests.
//
// If the test configuration is for a scale set, then each instance of that
// scale set counts as a separate test. Since we can't get the number of
// instances in each scale set we allow the `tests` attribute to be overridden
// in the usage file.
//
// Resource information: https://azure.microsoft.com/en-gb/pricing/details/network-watcher/#overview
// Pricing information: https://azure.microsoft.com/en-gb/pricing/details/network-watcher/#pricing
type NetworkConnectionMonitor struct {
	Address string
	Region  string

	Tests *int64 `infracost_usage:"tests"`
}

// CoreType returns the name of this resource type
func (r *NetworkConnectionMonitor) CoreType() string {
	return "NetworkConnectionMonitor"
}

// UsageSchema defines a list which represents the usage schema of NetworkConnectionMonitor.
func (r *NetworkConnectionMonitor) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "tests", DefaultValue: 0, ValueType: schema.Int64},
	}
}

// PopulateUsage parses the u schema.UsageData into the NetworkConnectionMonitor.
// It uses the `infracost_usage` struct tags to populate data into the NetworkConnectionMonitor.
func (r *NetworkConnectionMonitor) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid NetworkConnectionMonitor struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *NetworkConnectionMonitor) BuildResource() *schema.Resource {
	costComponents := r.testsCostComponents()

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *NetworkConnectionMonitor) testsCostComponents() []*schema.CostComponent {
	var qty *decimal.Decimal
	if r.Tests != nil {
		qty = decimalPtr(decimal.NewFromInt(*r.Tests))
	}

	tierData := []struct {
		suffix     string
		startUsage string
	}{
		{suffix: "(0-10)", startUsage: "0"},
		{suffix: "(10-240,010)", startUsage: "10"},
		{suffix: "(240,010-750,010)", startUsage: "240010"},
		{suffix: "(750,010-1,000,010)", startUsage: "750010"},
		{suffix: "(1,000,010+)", startUsage: "1000010"},
	}

	tierLimits := []int{10, 240000, 510000, 250000}

	var costComponents []*schema.CostComponent

	if len(tierData) == 0 {
		return costComponents
	}

	if qty == nil {
		costComponents = append(costComponents, r.buildTestsCostComponent(tierData[0].suffix, tierData[0].startUsage, nil))
	} else {
		tiers := usage.CalculateTierBuckets(*qty, tierLimits)
		for i, d := range tierData {
			// Skip the first tier since it's free
			if i == 0 {
				continue
			}

			if tiers[i].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, r.buildTestsCostComponent(d.suffix, d.startUsage, decimalPtr(tiers[i])))
			}
		}
	}

	return costComponents
}

func (r *NetworkConnectionMonitor) buildTestsCostComponent(suffix string, startUsage string, qty *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Tests %s", suffix),
		Unit:            "tests",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Network Watcher"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr("Standard Connection Monitor Test")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(startUsage),
		},
		UsageBased: true,
	}
}
