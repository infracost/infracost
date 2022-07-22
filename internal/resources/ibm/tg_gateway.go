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

func (r *TgGateway) connectionCostComponent(quantity *int64) *schema.CostComponent {
	var q *decimal.Decimal
	if quantity != nil {
		q = decimalPtr(decimal.NewFromInt(*quantity))
	}
	return &schema.CostComponent{
		Name:            "Connections",
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

func (r *TgGateway) dataTransferLocalCostComponent(quantity *float64) *schema.CostComponent {
	var q *decimal.Decimal
	if quantity != nil {
		q = decimalPtr(decimal.NewFromFloat(*quantity))
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

func (r *TgGateway) dataTransferGlobalCostComponent(quantity *float64) *schema.CostComponent {
	var q *decimal.Decimal
	if quantity != nil {
		q = decimalPtr(decimal.NewFromFloat(*quantity))
	}
	return &schema.CostComponent{
		Name:            "Data Transfer Global",
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
		r.connectionCostComponent(r.Connection),
	}

	if r.GlobalRouting {
		costComponents = append(costComponents, r.dataTransferGlobalCostComponent(r.DataTransferGlobal))
	} else {
		costComponents = append(costComponents, r.dataTransferLocalCostComponent(r.DataTransferLocal))
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    TgGatewayUsageSchema,
		CostComponents: costComponents,
	}
}
