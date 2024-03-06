package azure

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
)

// ServiceBusNamespace struct represents Azure Service Bus Namespace
//
// This resource is charged based on the SKU (Basic, Standard or Premium) and
// Capacity (only for Premium).
//
// Relay hours and Hybrid connection pricing should be associated with other
// Terraform resources in the future (azurerm_relay_namespace,
// azurerm_relay_hybrid_connection).
//
// Resource information: https://azure.microsoft.com/en-gb/pricing/details/service-bus/#pricing
// Pricing information: https://azure.microsoft.com/en-gb/pricing/details/service-bus/#pricing
type ServiceBusNamespace struct {
	Address  string
	Region   string
	SKU      string
	Capacity int64

	// Usage-based fields
	MonthlyMessagingOperations *int64 `infracost_usage:"monthly_messaging_operations"`
	MonthlyBrokeredConnections *int64 `infracost_usage:"monthly_brokered_connections"`
}

// CoreType returns the name of this resource type
func (r *ServiceBusNamespace) CoreType() string {
	return "ServiceBusNamespace"
}

// UsageSchema defines a list which represents the usage schema of ServiceBusNamespace.
func (r *ServiceBusNamespace) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_messaging_operations", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_brokered_connections", DefaultValue: 0, ValueType: schema.Int64},
	}
}

// PopulateUsage parses the u schema.UsageData into the ServiceBusNamespace.
// It uses the `infracost_usage` struct tags to populate data into the ServiceBusNamespace.
func (r *ServiceBusNamespace) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid ServiceBusNamespace struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *ServiceBusNamespace) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	if strings.EqualFold(r.SKU, "premium") {
		costComponents = append(costComponents, r.messagingUnitsCostComponent())
	} else if strings.EqualFold(r.SKU, "basic") {
		costComponents = append(costComponents, r.messagingOperationsCostComponents()...)
	} else { // standard
		costComponents = append(costComponents, r.baseChargeCostComponent())
		costComponents = append(costComponents, r.messagingOperationsCostComponents()...)
		costComponents = append(costComponents, r.brokeredConnectionsCostComponents()...)
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *ServiceBusNamespace) baseChargeCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Base charge",
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Service Bus"),
			ProductFamily: strPtr("Integration"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(r.normalizedSku())},
				{Key: "meterName", ValueRegex: regexPtr("Base Unit$")},
			},
		},
	}
}

func (r *ServiceBusNamespace) messagingOperationsCostComponents() []*schema.CostComponent {
	var qty *decimal.Decimal
	if r.MonthlyMessagingOperations != nil {
		qty = decimalPtr(decimal.NewFromInt(*r.MonthlyMessagingOperations))
	}

	if strings.EqualFold(r.SKU, "basic") {
		return []*schema.CostComponent{
			r.buildMessagingOperationsCostComponent("", "0", qty),
		}
	}

	tierData := []struct {
		suffix     string
		startUsage string
	}{
		{suffix: " (first 13M)", startUsage: "0"},
		{suffix: " (13M-100M)", startUsage: "13"},
		{suffix: " (100M-2,500M)", startUsage: "100"},
		{suffix: " (over 2,500M)", startUsage: "2500"},
	}

	tierLimits := []int{13_000_000, 87_000_000, 2_400_000_000}

	var costComponents []*schema.CostComponent

	if qty == nil {
		costComponents = append(costComponents, r.buildMessagingOperationsCostComponent(tierData[1].suffix, tierData[1].startUsage, nil))
	} else {
		tiers := usage.CalculateTierBuckets(*qty, tierLimits)
		for i, d := range tierData {
			// Skip the first tier since it's free
			if i == 0 {
				continue
			}

			if tiers[i].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, r.buildMessagingOperationsCostComponent(d.suffix, d.startUsage, decimalPtr(tiers[i])))
			}
		}
	}

	return costComponents
}

func (r *ServiceBusNamespace) buildMessagingOperationsCostComponent(suffix, startUsage string, qty *decimal.Decimal) *schema.CostComponent {
	var perMQty *decimal.Decimal
	if qty != nil {
		perMQty = decimalPtr(qty.Div(decimal.NewFromInt(1_000_000)))
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Messaging operations%s", suffix),
		Unit:            "1M operations",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: perMQty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Service Bus"),
			ProductFamily: strPtr("Integration"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(r.normalizedSku())},
				{Key: "meterName", ValueRegex: regexPtr("Messaging Operations$")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(startUsage),
		},
		UsageBased: true,
	}
}

func (r *ServiceBusNamespace) messagingUnitsCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Messaging units",
		Unit:           "units",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(r.Capacity)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Service Bus"),
			ProductFamily: strPtr("Integration"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(r.normalizedSku())},
				{Key: "meterName", ValueRegex: regexPtr("Messaging Unit$")},
			},
		},
	}
}

func (r *ServiceBusNamespace) brokeredConnectionsCostComponents() []*schema.CostComponent {
	var qty *decimal.Decimal
	if r.MonthlyBrokeredConnections != nil {
		qty = decimalPtr(decimal.NewFromInt(*r.MonthlyBrokeredConnections))
	}

	tierData := []struct {
		suffix     string
		startUsage string
	}{
		{suffix: " (first 1K)", startUsage: "0"},
		{suffix: " (1K-100K)", startUsage: "1000"},
		{suffix: " (100K-500K)", startUsage: "500000"},
		{suffix: " (over 500K)", startUsage: "100000"},
	}

	tierLimits := []int{1000, 99000, 400000}

	var costComponents []*schema.CostComponent

	if qty == nil {
		costComponents = append(costComponents, r.buildBrokeredConnectionsCostComponent(tierData[1].suffix, tierData[1].startUsage, nil))
	} else {
		tiers := usage.CalculateTierBuckets(*qty, tierLimits)
		for i, d := range tierData {
			// Skip the first tier since it's free
			if i == 0 {
				continue
			}

			if tiers[i].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, r.buildBrokeredConnectionsCostComponent(d.suffix, d.startUsage, decimalPtr(tiers[i])))
			}
		}
	}

	return costComponents
}

func (r *ServiceBusNamespace) buildBrokeredConnectionsCostComponent(suffix, startUsage string, qty *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Brokered connections%s", suffix),
		Unit:            "connections",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Service Bus"),
			ProductFamily: strPtr("Integration"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(r.normalizedSku())},
				{Key: "meterName", ValueRegex: regexPtr("Brokered Connection$")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(startUsage),
		},
		UsageBased: true,
	}
}

func (r *ServiceBusNamespace) normalizedSku() string {
	return cases.Title(language.English).String(strings.ToLower(r.SKU))
}
