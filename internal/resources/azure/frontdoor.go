package azure

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
)

// Frontdoor struct represents Azure's Front Door network service.
//
// More resource information here: https://docs.microsoft.com/en-us/azure/frontdoor/front-door-overview
// Pricing information here: https://azure.microsoft.com/en-us/pricing/details/frontdoor/ (Azure Front Door tab)
type Frontdoor struct {
	Address string
	Region  string

	FrontendHosts int
	RoutingRules  int

	// "usage" args
	MonthlyInboundDataTransferGB  *float64                            `infracost_usage:"monthly_inbound_data_transfer_gb"`
	MonthlyOutboundDataTransferGB *frontdoorOutboundDataTransferUsage `infracost_usage:"monthly_outbound_data_transfer_gb"`
}

// frontdoorOutboundDataTransferUsage represents outbound transfer usage group per
// region.
type frontdoorOutboundDataTransferUsage struct {
	USGovZone1MonthlyTransferGB *float64 `infracost_usage:"us_gov"`
	Zone1MonthlyTransferGB      *float64 `infracost_usage:"north_america_europe_africa"`
	Zone2MonthlyTransferGB      *float64 `infracost_usage:"asia_pacific"`
	Zone3MonthlyTransferGB      *float64 `infracost_usage:"south_america"`
	Zone4MonthlyTransferGB      *float64 `infracost_usage:"australia"`
	Zone5MonthlyTransferGB      *float64 `infracost_usage:"india"`
}

// CoreType returns the name of this resource type
func (r *Frontdoor) CoreType() string {
	return "Frontdoor"
}

// UsageSchema defines a list which represents the usage schema of EventGridTopic.
func (r *Frontdoor) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_inbound_data_transfer_gb", DefaultValue: 0, ValueType: schema.Float64},
		{
			Key:          "monthly_outbound_data_transfer_gb",
			DefaultValue: &usage.ResourceUsage{Name: "monthly_outbound_data_transfer_gb", Items: frontdoorOutboundDataUsageSchema},
			ValueType:    schema.SubResourceUsage,
		},
	}
}

// frontdoorOutboundDataUsageSchema defines a nested list of outbound data
// transfer usage per region.
var frontdoorOutboundDataUsageSchema = []*schema.UsageItem{
	{Key: "us_gov", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "north_america_europe_africa", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "asia_pacific", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "south_america", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "australia", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "india", DefaultValue: 0, ValueType: schema.Float64},
}

// PopulateUsage parses the u schema.UsageData into the Frontdoor.
// It uses the `infracost_usage` struct tags to populate data into the Frontdoor.
func (r *Frontdoor) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from valid Frontdoor data.
// This method is called after the resource is initialised by an IaC provider.
func (r *Frontdoor) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	costComponents = append(costComponents, r.routingRulesCostComponents()...)
	costComponents = append(costComponents, r.frontendHostsCostComponents()...)
	costComponents = append(costComponents, r.inboundDataTransferCostComponents()...)

	// Subresource is used because the cost component has nested tier items
	outboundTransferSubResource := &schema.Resource{
		Name:           "Outbound data transfer",
		CostComponents: r.outboundDataTransferCostComponents(),
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
		SubResources:   []*schema.Resource{outboundTransferSubResource},
	}
}

// routingRulesCostComponents returns cost components for defined routing rules.
// The pricing depends on rules quantity.
func (r *Frontdoor) routingRulesCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	name := "Routing rules"
	firstBatchThreshold := 5
	quantity := r.RoutingRules

	type dataTier struct {
		name       string
		meterName  string
		startUsage string
	}

	// For Zones 1 - 5 the price with startUsage 0 in API is free, but Azure
	// pricing calculator bills first rules.
	// US Gov Zone 1 has only one price
	firstRulesStartUsage := "5"
	if r.isGovZone(r.Region) {
		firstRulesStartUsage = "0"
	}

	data := []dataTier{
		{
			name:       fmt.Sprintf("%s (first %d rules)", name, firstBatchThreshold),
			meterName:  "Included Routing Rules",
			startUsage: firstRulesStartUsage,
		}, {
			name:       fmt.Sprintf("%s (per additional rule)", name),
			meterName:  "Overage Routing Rules",
			startUsage: "0",
		},
	}

	tierLimits := []int{firstBatchThreshold}
	tiers := usage.CalculateTierBuckets(decimal.NewFromInt(int64(quantity)), tierLimits)
	for i, d := range data {
		if i < len(tiers) && tiers[i].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, &schema.CostComponent{
				Name:           d.name,
				Unit:           "hours",
				UnitMultiplier: decimal.NewFromInt(1),
				HourlyQuantity: decimalPtr(tiers[i]),
				ProductFilter:  r.buildProductFilter(d.meterName),
				PriceFilter: &schema.PriceFilter{
					PurchaseOption:   strPtr("Consumption"),
					StartUsageAmount: strPtr(d.startUsage),
				},
			})
		}
	}

	return costComponents
}

// frontendHostsCostComponents returns a cost component for frontend hosts.
// Only additional hosts above free limit are billed.
func (r *Frontdoor) frontendHostsCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	freeQuantity := 100
	quantity := r.FrontendHosts

	if quantity <= freeQuantity {
		return costComponents
	}

	name := fmt.Sprintf("Frontend hosts (over %d)", freeQuantity)
	billedHostsQuantity := quantity - freeQuantity

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            name,
		Unit:            "hosts",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(int64(billedHostsQuantity))),
		ProductFilter:   r.buildProductFilter("Custom Domain"),
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	})

	return costComponents
}

// inboundDataTransferCostComponents returns a cost component for amount of
// transferred inbound data.
func (r *Frontdoor) inboundDataTransferCostComponents() []*schema.CostComponent {
	return []*schema.CostComponent{
		{
			Name:            "Inbound data transfer",
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: floatPtrToDecimalPtr(r.MonthlyInboundDataTransferGB),
			ProductFilter:   r.buildProductFilter("Data Transfer In"),
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("Consumption"),
			},
			UsageBased: true,
		},
	}
}

// outboundDataTransferCostComponents returns cost components for amount of
// transferred outbound data. There are several tiers that are billed
// differently. The pricing depends on the region.
func (r *Frontdoor) outboundDataTransferCostComponents() []*schema.CostComponent {
	if r.isGovZone(r.Region) {
		return r.govOutboundDataTransferCostComponents()
	}

	costComponents := []*schema.CostComponent{}

	resourceUsage := &frontdoorOutboundDataTransferUsage{}
	if r.MonthlyOutboundDataTransferGB != nil {
		resourceUsage = r.MonthlyOutboundDataTransferGB
	}

	type zoneUsage struct {
		zone     string
		name     string
		quantity *float64
	}

	zones := []zoneUsage{
		{zone: "Zone 1", quantity: resourceUsage.Zone1MonthlyTransferGB, name: "North America, Europe and Africa"},
		{zone: "Zone 2", quantity: resourceUsage.Zone2MonthlyTransferGB, name: "Asia Pacific (including Japan)"},
		{zone: "Zone 3", quantity: resourceUsage.Zone3MonthlyTransferGB, name: "South America"},
		{zone: "Zone 4", quantity: resourceUsage.Zone4MonthlyTransferGB, name: "Australia"},
		{zone: "Zone 5", quantity: resourceUsage.Zone5MonthlyTransferGB, name: "India"},
	}

	currentZone := zones[0]
	for _, item := range zones {
		if strings.EqualFold(item.zone, r.Region) {
			currentZone = item
			break
		}
	}

	type dataTier struct {
		name       string
		startUsage string
	}
	data := []dataTier{
		{name: fmt.Sprintf("%s (first 10TB)", currentZone.name), startUsage: "0"},
		{name: fmt.Sprintf("%s (next 40TB)", currentZone.name), startUsage: "10000"},
		{name: fmt.Sprintf("%s (over 50TB)", currentZone.name), startUsage: "50000"},
	}

	if currentZone.quantity != nil {
		quantity := decimal.NewFromFloat(*currentZone.quantity)

		tierLimits := []int{10000, 40000}
		tiers := usage.CalculateTierBuckets(quantity, tierLimits)

		for i, d := range data {
			if i < len(tiers) && tiers[i].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, r.buildOutboundDataTransferCostComponent(
					d.name,
					d.startUsage,
					decimalPtr(tiers[i]),
				))
			}
		}
	} else {
		costComponents = append(costComponents, r.buildOutboundDataTransferCostComponent(
			data[0].name,
			data[0].startUsage,
			nil,
		))
	}
	return costComponents
}

// govOutboundDataTransferCostComponents returns a cost component for outbound
// data transfer in US Gov zone. This zone doesn't have tiers.
func (r *Frontdoor) govOutboundDataTransferCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	var quantity *decimal.Decimal

	if r.MonthlyOutboundDataTransferGB != nil {
		transferAmount := r.MonthlyOutboundDataTransferGB.USGovZone1MonthlyTransferGB
		if transferAmount != nil {
			quantity = decimalPtr(decimal.NewFromFloat(*transferAmount))
		}
	}

	costComponents = append(costComponents, r.buildOutboundDataTransferCostComponent(
		"US Gov",
		"0",
		quantity,
	))

	return costComponents
}

// buildOutboundDataTransferCostComponent returns a cost component for one tier
// of outbound data transfer usage.
func (r *Frontdoor) buildOutboundDataTransferCostComponent(name, startUsage string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter:   r.buildProductFilter("Data Transfer Out"),
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr(startUsage),
		},
		UsageBased: true,
	}
}

// buildProductFilter returns a product filter for the Front Door's products.
//
// skuName and productName define the original Front Door service (not
// Standard/Premium).
func (r *Frontdoor) buildProductFilter(meterName string) *schema.ProductFilter {
	return &schema.ProductFilter{
		VendorName:    strPtr("azure"),
		Region:        strPtr(r.Region),
		Service:       strPtr("Azure Front Door Service"),
		ProductFamily: strPtr("Networking"),
		AttributeFilters: []*schema.AttributeFilter{
			{Key: "skuName", Value: strPtr("Standard")},
			{Key: "meterName", ValueRegex: regexPtr(fmt.Sprintf("%s$", meterName))},
			{Key: "productName", Value: strPtr("Azure Front Door Service")},
		},
	}
}

// isGovZone checks if the region/zone is US Gov
func (r *Frontdoor) isGovZone(zone string) bool {
	return strings.EqualFold(zone, "US Gov Zone 1")
}
