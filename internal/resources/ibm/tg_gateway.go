package ibm

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// TgGateway struct represents a transit gateway between VPCs
//
// Resource information: https://cloud.ibm.com/docs/transit-gateway?topic=transit-gateway-getting-started
// Pricing information: https://cloud.ibm.com/docs/transit-gateway?topic=transit-gateway-tg-pricing
type TgGateway struct {
	Address       string
	Region        string
	GlobalRouting bool

	DataTransferLocal  *float64 `infracost_usage:"data_transfer_local"`
	DataTransferGlobal *float64 `infracost_usage:"data_transfer_global"`
	Connection         *int64   `infracost_usage:"connection"`
}

// TgGatewayUsageSchema defines a list which represents the usage schema of TgGateway.
var TgGatewayUsageSchema = []*schema.UsageItem{
	{Key: "data_transfer_local", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "data_transfer_global", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "connection", DefaultValue: 0, ValueType: schema.Float64},
}

// PopulateUsage parses the u schema.UsageData into the TgGateway.
// It uses the `infracost_usage` struct tags to populate data into the TgGateway.
func (r *TgGateway) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *TgGateway) connectionFreeCostComponent() *schema.CostComponent {
	var q *decimal.Decimal
	if r.Connection != nil {
		q = decimalPtr(decimal.NewFromInt(*r.Connection))
		if q.GreaterThan(decimal.NewFromInt(10)) {
			q = decimalPtr(decimal.NewFromInt(10))
		}
	}
	component := schema.CostComponent{
		Name:            "Connections Free allowance (first 10 Connection)",
		Unit:            "Connection",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Service:       strPtr("transit.gateway"),
			ProductFamily: strPtr("service"),
			Region:        strPtr("global"),
		},
	}
	component.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
	return &component
}

func (r *TgGateway) connectionCostComponent() *schema.CostComponent {
	var q *decimal.Decimal
	if r.Connection != nil {
		q = decimalPtr(decimal.NewFromInt(*r.Connection))
		if q.LessThanOrEqual(decimal.NewFromInt(10)) {
			q = decimalPtr(decimal.NewFromInt(0))
		} else {
			q = decimalPtr(q.Sub(decimal.NewFromInt(10)))
		}
	}
	return &schema.CostComponent{
		Name:            "Additional Connections",
		Unit:            "Connection",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Service:       strPtr("transit.gateway"),
			ProductFamily: strPtr("service"),
			Region:        strPtr("global"),
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("INSTANCES"),
		},
	}
}

func (r *TgGateway) dataTransferLocalCostComponent() *schema.CostComponent {
	var q *decimal.Decimal
	if r.DataTransferLocal != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.DataTransferLocal))
	}
	return &schema.CostComponent{
		Name:            "Data Transfer Local",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Service:       strPtr("transit.gateway"),
			ProductFamily: strPtr("service"),
			Region:        strPtr("global"),
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("GIGABYTE_TRANSMITTEDS_LOCAL"),
		},
	}
}

func (r *TgGateway) dataTransferGlobalFreeCostComponent() *schema.CostComponent {
	var q *decimal.Decimal
	if r.DataTransferGlobal != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.DataTransferGlobal))
	}
	if q.GreaterThan(decimal.NewFromInt(1000)) {
		q = decimalPtr(decimal.NewFromInt(1000))
	}
	component := schema.CostComponent{
		Name:            "Data Transfer Global Free allowance (first 1000 GB)",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
	}
	component.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
	return &component
}

func (r *TgGateway) dataTransferGlobalCostComponent() *schema.CostComponent {
	var q *decimal.Decimal
	if r.DataTransferGlobal != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.DataTransferGlobal))
	}
	if q.LessThanOrEqual(decimal.NewFromInt(1000)) {
		q = decimalPtr(decimal.NewFromInt(0))
	} else {
		q = decimalPtr(q.Sub(decimal.NewFromInt(1000)))
	}
	return &schema.CostComponent{
		Name:            "Data Transfer Global (over 1000 GB)",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Service:       strPtr("transit.gateway"),
			ProductFamily: strPtr("service"),
			Region:        strPtr("global"),
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("GIGABYTE_TRANSMITTEDS_GLOBAL"),
		},
	}
}

// BuildResource builds a schema.Resource from a valid TgGateway struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *TgGateway) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.connectionFreeCostComponent(),
		r.connectionCostComponent(),
	}

	if r.GlobalRouting {
		costComponents = append(costComponents, r.dataTransferGlobalFreeCostComponent())
		costComponents = append(costComponents, r.dataTransferGlobalCostComponent())
	} else {
		costComponents = append(costComponents, r.dataTransferLocalCostComponent())
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    TgGatewayUsageSchema,
		CostComponents: costComponents,
	}
}
