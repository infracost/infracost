package azure

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// LogicAppStandard struct represents Azure Logic App Standard.
//
// This resource's pricing is based on the SKU size and the number of standard
// and enterprise connector calls. The SKU size is determined by the related
// App Service Plan SKU, but can be overwritten with a usage-based attribute
// in the usage file. If the SKU cannot be determined we show the pricing per
// vCore and GB of memory. This resource only supports the Standard Plan pricing,
// not the Consumption Plan.
//
// Resource information: https://azure.microsoft.com/en-gb/pricing/details/logic-apps/
// Pricing information: https://azure.microsoft.com/en-gb/pricing/details/logic-apps/#pricing
type LogicAppStandard struct {
	Address string
	Region  string

	// Usage-based attributes
	// This is set from the related App Service Plan SKU, but can be overwritten in the usage file
	SKU                             *string `infracost_usage:"sku"`
	MonthlyStandardConnectorCalls   *int64  `infracost_usage:"monthly_standard_connector_calls"`
	MonthlyEnterpriseConnectorCalls *int64  `infracost_usage:"monthly_enterprise_connector_calls"`
}

var logicAppSKUResources = map[string]struct {
	vCores decimal.Decimal
	memory decimal.Decimal
}{
	"WS1": {vCores: decimal.NewFromInt(1), memory: decimal.NewFromFloat(3.5)},
	"WS2": {vCores: decimal.NewFromInt(2), memory: decimal.NewFromFloat(7)},
	"WS3": {vCores: decimal.NewFromInt(4), memory: decimal.NewFromFloat(14)},
}

// CoreType returns the name of this resource type
func (r *LogicAppStandard) CoreType() string {
	return "LogicAppStandard"
}

// UsageSchema defines a list which represents the usage schema of LogicAppStandard.
func (r *LogicAppStandard) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "sku", DefaultValue: "", ValueType: schema.String},
		{Key: "monthly_standard_connector_calls", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_enterprise_connector_calls", DefaultValue: 0, ValueType: schema.Int64},
	}
}

// PopulateUsage parses the u schema.UsageData into the LogicAppStandard.
// It uses the `infracost_usage` struct tags to populate data into the LogicAppStandard.
func (r *LogicAppStandard) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid LogicAppStandard struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *LogicAppStandard) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.workflowVCoreCostComponent(),
		r.workflowMemoryCostComponent(),
		r.standardConnectorCostComponent(),
		r.enterpriseConnectorCostComponent(),
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *LogicAppStandard) workflowVCoreCostComponent() *schema.CostComponent {
	sku := r.normalizedSKU()

	var qty *decimal.Decimal
	if r, ok := logicAppSKUResources[sku]; ok {
		qty = decimalPtr(r.vCores)
	}

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Workflow vCore (%s)", sku),
		Unit:           "vCore",
		HourlyQuantity: qty,
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Logic Apps"),
			ProductFamily: strPtr("Integration"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr("Standard vCPU Duration")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func (r *LogicAppStandard) workflowMemoryCostComponent() *schema.CostComponent {
	sku := r.normalizedSKU()

	var qty *decimal.Decimal
	if r, ok := logicAppSKUResources[sku]; ok {
		qty = decimalPtr(r.memory)
	}

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Workflow memory (%s)", sku),
		Unit:           "GB",
		HourlyQuantity: qty,
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Logic Apps"),
			ProductFamily: strPtr("Integration"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr("Standard Memory Duration")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func (r *LogicAppStandard) standardConnectorCostComponent() *schema.CostComponent {
	var qty *decimal.Decimal
	if r.MonthlyStandardConnectorCalls != nil {
		qty = decimalPtr(decimal.NewFromInt(*r.MonthlyStandardConnectorCalls))
	}

	return &schema.CostComponent{
		Name:            "Standard connectors",
		Unit:            "calls",
		MonthlyQuantity: qty,
		UnitMultiplier:  decimal.NewFromInt(1),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Logic Apps"),
			ProductFamily: strPtr("Integration"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr("Consumption Standard Connector Actions")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
		UsageBased: true,
	}
}

func (r *LogicAppStandard) enterpriseConnectorCostComponent() *schema.CostComponent {
	var qty *decimal.Decimal
	if r.MonthlyEnterpriseConnectorCalls != nil {
		qty = decimalPtr(decimal.NewFromInt(*r.MonthlyEnterpriseConnectorCalls))
	}

	return &schema.CostComponent{
		Name:            "Enterprise connectors",
		Unit:            "calls",
		MonthlyQuantity: qty,
		UnitMultiplier:  decimal.NewFromInt(1),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Logic Apps"),
			ProductFamily: strPtr("Integration"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr("Consumption Enterprise Connector Actions")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
		UsageBased: true,
	}
}

func (r *LogicAppStandard) normalizedSKU() string {
	if r.SKU != nil {
		skuName := strings.ToUpper(*r.SKU)

		if _, ok := logicAppSKUResources[skuName]; ok {
			return skuName
		}
	}

	return "unknown SKU"
}
