package ibm

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// Graduated Tier pricing model
const DNS_SERVICES_PROGRAMMATIC_PLAN_NAME string = "standard-dns"

func GetDNSServicesCostComponents(r *ResourceInstance) []*schema.CostComponent {
	if r.Plan == DNS_SERVICES_PROGRAMMATIC_PLAN_NAME {
		return []*schema.CostComponent{
			DNSServicesZonesCostComponent(r),
			DNSServicesPoolsPerHourCostComponent(r),
			DNSServicesGLBInstancesPerHourCostComponent(r),
			DNSServicesHealthChecksCostComponent(r),
			DNSServicesCustomResolverLocationsPerHourCostComponent(r),
			DNSServicesMillionCustomResolverExternalQueriesCostComponent(r),
			DNSServicesMillionDNSQueriesCostComponent(r),
		}
	} else {
		costComponent := schema.CostComponent{
			Name:            fmt.Sprintf("Plan %s with customized pricing", r.Plan),
			UnitMultiplier:  decimal.NewFromInt(1), // Final quantity for this cost component will be divided by this amount
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		}
		costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
		return []*schema.CostComponent{
			&costComponent,
		}
	}
}

// Unit: ITEMS (Linear Tier)
func DNSServicesZonesCostComponent(r *ResourceInstance) *schema.CostComponent {

	var zones_included int = 1
	var quantity *decimal.Decimal

	if r.DNSServices_Zones != nil {
		additional_zones := *r.DNSServices_Zones - int64(zones_included)
		if additional_zones > 0 {
			quantity = decimalPtr(decimal.NewFromInt(additional_zones))
		} else {
			quantity = decimalPtr(decimal.NewFromInt(0))
		}
	}

	costComponent := schema.CostComponent{
		Name:            "Additional Zones",
		Unit:            "Zones",
		UnitMultiplier:  decimal.NewFromFloat(1), // Final quantity for this cost component will be divided by this amount
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("ITEMS"),
		},
	}
	return &costComponent
}

// Unit: NUMBERPOOLS (Linear Tier)
func DNSServicesPoolsPerHourCostComponent(r *ResourceInstance) *schema.CostComponent {

	var quantity *decimal.Decimal

	if (r.DNSServices_PoolHours != nil) && (r.DNSServices_Pools != nil) {
		quantity = decimalPtr(decimal.NewFromFloat(*r.DNSServices_PoolHours * float64(*r.DNSServices_Pools)))
	}

	costComponent := schema.CostComponent{
		Name:            "Pool Hours",
		Unit:            "Hours",
		UnitMultiplier:  decimal.NewFromFloat(1), // Final quantity for this cost component will be divided by this amount
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("NUMBERPOOLS"),
		},
	}
	return &costComponent
}

// Unit: NUMBERGLB (Linear Tier)
func DNSServicesGLBInstancesPerHourCostComponent(r *ResourceInstance) *schema.CostComponent {

	var quantity *decimal.Decimal

	if (r.DNSServices_GLBInstanceHours != nil) && (r.DNSServices_GLBInstances != nil) {
		quantity = decimalPtr(decimal.NewFromFloat(*r.DNSServices_GLBInstanceHours * float64(*r.DNSServices_GLBInstances)))
	}

	costComponent := schema.CostComponent{
		Name:            "GLB Instance Hours",
		Unit:            "Hours",
		UnitMultiplier:  decimal.NewFromFloat(1), // Final quantity for this cost component will be divided by this amount
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("NUMBERGLB"),
		},
	}
	return &costComponent
}

// Unit: NUMBERHEALTHCHECK (Linear Tier)
func DNSServicesHealthChecksCostComponent(r *ResourceInstance) *schema.CostComponent {

	var quantity *decimal.Decimal

	if r.DNSServices_HealthChecks != nil {
		quantity = decimalPtr(decimal.NewFromInt(*r.DNSServices_HealthChecks))
	}

	costComponent := schema.CostComponent{
		Name:            "Health Checks",
		Unit:            "Health Checks",
		UnitMultiplier:  decimal.NewFromFloat(1), // Final quantity for this cost component will be divided by this amount
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("NUMBERHEALTHCHECK"),
		},
	}
	return &costComponent
}

// Unit: RESOLVERLOCATIONS (Linear Tier)
func DNSServicesCustomResolverLocationsPerHourCostComponent(r *ResourceInstance) *schema.CostComponent {

	var quantity *decimal.Decimal

	if (r.DNSServices_CustomResolverLocationHours != nil) && (r.DNSServices_CustomResolverLocations != nil) {
		quantity = decimalPtr(decimal.NewFromFloat(*r.DNSServices_CustomResolverLocationHours * float64(*r.DNSServices_CustomResolverLocations)))
	}

	costComponent := schema.CostComponent{
		Name:            "Custom Resolver Location Hours",
		Unit:            "Hours",
		UnitMultiplier:  decimal.NewFromFloat(1), // Final quantity for this cost component will be divided by this amount
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("RESOLVERLOCATIONS"),
		},
	}
	return &costComponent
}

// Unit: MILLION_ITEMS_CREXTERNALQUERIES (Graduated Tier)
func DNSServicesMillionCustomResolverExternalQueriesCostComponent(r *ResourceInstance) *schema.CostComponent {

	var quantity *decimal.Decimal

	if r.DNSServices_CustomResolverExternalQueries != nil {
		quantity = decimalPtr(decimal.NewFromInt(*r.DNSServices_CustomResolverExternalQueries))
	}

	costComponent := schema.CostComponent{
		Name:            "Million Custom Resolver External Queries",
		Unit:            "Million Queries",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("MILLION_ITEMS_CREXTERNALQUERIES"),
		},
	}
	return &costComponent
}

// Unit: MILLION_ITEMS (Graduated Tier)
func DNSServicesMillionDNSQueriesCostComponent(r *ResourceInstance) *schema.CostComponent {

	var million_dns_queries_included float32 = 1
	var quantity *decimal.Decimal

	if r.DNSServices_DNSQueries != nil {
		additional_million_dns_queries := *r.DNSServices_DNSQueries - int64(million_dns_queries_included)
		if additional_million_dns_queries > 0 {
			quantity = decimalPtr(decimal.NewFromInt(additional_million_dns_queries))
		} else {
			quantity = decimalPtr(decimal.NewFromInt(0))
		}
	}

	costComponent := schema.CostComponent{
		Name:            "Additional Million DNS Queries",
		Unit:            "Million Queries",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("MILLION_ITEMS"),
		},
	}
	return &costComponent
}
