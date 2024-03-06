package azure

import (
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// EventGridTopic struct represents an Azure Event Grid Topic, a fully managed
// event routing service that simplifies the process of creating and managing
// event-driven applications.
//
// Azure Event Grid allows you to build reactive applications by reacting to
// events from various Azure services, custom sources, or on-premises
// infrastructure.
//
// EventGridTopic is used for both System (azurerm_eventgrid_system_topic) and
// Custom (azurerm_eventgrid_topic) topics.
//
// System Topics are predefined, multi-tenant topics that are built-in to Azure
// services and emit events directly from the service. Custom Topics are
// application and solution-specific topics that you define for your own
// applications to publish events to.
//
// For more information about Azure Event Grid System Topics and pricing, refer
// to the following links:
//
// System topic information: https://docs.microsoft.com/en-us/azure/event-grid/system-topics
// Custom topic information: https://learn.microsoft.com/en-us/azure/event-grid/custom-topics
// Pricing information: https://azure.microsoft.com/en-us/pricing/details/event-grid/
type EventGridTopic struct {
	Address           string
	Region            string
	MonthlyOperations *float64 `infracost_usage:"monthly_operations"`
}

// CoreType returns the name of this resource type
func (r *EventGridTopic) CoreType() string {
	return "EventGridTopic"
}

// UsageSchema defines a list which represents the usage schema of EventGridTopic.
func (r *EventGridTopic) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{
			Key:       "monthly_operations",
			ValueType: schema.Float64,
		},
	}
}

// PopulateUsage parses the u schema.UsageData into the EventGridTopic.
// It uses the `infracost_usage` struct tags to populate data into the EventGridTopic.
func (r *EventGridTopic) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid EventGridTopic struct.
// This method is called after the resource is initialized by an IAC provider.
// See providers folder for more information.
//
// The returned resource includes a CostComponent for Event Grid operations,
// taking into account the user's specified number of monthly operations. Azure
// Event Grid pricing is based on the number of operations, where each operation
// is defined as an event ingress, delivery attempt, or management call. The
// pricing is tiered, with the first 100,000 operations free, and then billed per
// 100k operations thereafter.
//
// Note: The pricing page for Azure Event Grid mistakenly describes that it is
// billed per million operations. This is incorrect and has been verified by the
// https://azure.microsoft.com/en-us/pricing/calculator/ and information in the
// cloud pricing API.
func (r *EventGridTopic) BuildResource() *schema.Resource {
	var quantity *decimal.Decimal
	if r.MonthlyOperations != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.MonthlyOperations).Div(decimal.NewFromInt(100000)).RoundDown(0))
	}

	costComponents := []*schema.CostComponent{
		{
			Name:            "Operations",
			Unit:            "100k operations",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: quantity,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr(vendorName),
				Region:        strPtr(r.Region),
				Service:       strPtr("Event Grid"),
				ProductFamily: strPtr("Internet of Things"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "skuName", Value: strPtr("Standard")},
					{Key: "meterName", Value: strPtr("Standard Operations")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption:   strPtr("Consumption"),
				StartUsageAmount: strPtr("1"),
			},
			UsageBased: true,
		},
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}
