package azure

import (
	"fmt"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// TrafficManagerEndpoint struct represents Azure Traffic Manager Endpoints.
//
// Resource information: https://learn.microsoft.com/en-us/azure/traffic-manager/traffic-manager-endpoint-types
// Pricing information: https://azure.microsoft.com/en-us/pricing/details/traffic-manager/#pricing
type TrafficManagerEndpoint struct {
	Address string
	Region  string

	ProfileEnabled      bool
	External            bool
	HealthCheckInterval int64
}

// CoreType returns the name of this resource type
func (r *TrafficManagerEndpoint) CoreType() string {
	return "TrafficManagerEndpoint"
}

// UsageSchema defines a list which represents the usage schema of TrafficManagerEndpoint.
func (r *TrafficManagerEndpoint) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

// PopulateUsage parses the u schema.UsageData into the TrafficManagerEndpoint.
// It uses the `infracost_usage` struct tags to populate data into the TrafficManagerEndpoint.
func (r *TrafficManagerEndpoint) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid TrafficManagerEndpoint struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *TrafficManagerEndpoint) BuildResource() *schema.Resource {
	if !r.ProfileEnabled {
		return &schema.Resource{
			Name: r.Address,
		}
	}

	costComponents := []*schema.CostComponent{
		r.healthCheckCostComponent(),
	}

	if r.HealthCheckInterval < 30 {
		costComponents = append(costComponents, r.fastHealthCheckCostComponent())
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *TrafficManagerEndpoint) sku() string {
	if r.External {
		return "Non-Azure Endpoint"
	} else {
		return "Azure Endpoint"
	}
}

func (r *TrafficManagerEndpoint) healthCheckCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Basic health check (%s)", trafficManagerBillingRegion(r.Region)),
		Unit:            "hours",
		UnitMultiplier:  schema.MonthToHourUnitMultiplier,
		UnitRounding:    int32Ptr(0),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),

		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(trafficManagerBillingRegion(r.Region)),
			Service:       strPtr("Traffic Manager"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(r.sku())},
				{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Health Checks", r.sku()))},
			},
		},
	}
}

func (r *TrafficManagerEndpoint) fastHealthCheckCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Fast interval health checks add-on (%s)", trafficManagerBillingRegion(r.Region)),
		Unit:            "hours",
		UnitMultiplier:  schema.MonthToHourUnitMultiplier,
		UnitRounding:    int32Ptr(0),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(trafficManagerBillingRegion(r.Region)),
			Service:       strPtr("Traffic Manager"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(r.sku())},
				{Key: "meterName", Value: strPtr(fmt.Sprintf("%s Fast Interval Health Check Add-ons", r.sku()))},
			},
		},
	}
}
