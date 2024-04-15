package ibm

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetSysdigTimeseriesCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.Monitoring_TimeSeriesHour != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.Monitoring_TimeSeriesHour))
	}
	return &schema.CostComponent{
		Name:            "Additional Time series",
		Unit:            "Time series hour",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("TIME_SERIES_HOURS"),
		},
	}
}

func GetSysdigContainerCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.Monitoring_ContainerHour != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.Monitoring_ContainerHour))
	}
	return &schema.CostComponent{
		Name:            "Additional Container Hours",
		Unit:            "Container Hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("CONTAINER_HOURS"),
		},
	}
}

func GetSysdigApiCallCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.Monitoring_APICall != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.Monitoring_APICall))
	}
	return &schema.CostComponent{
		Name:            "Additional API Calls",
		Unit:            "API Calls",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("API_CALL_HOURS"),
		},
	}
}

func GetSysdigNodeHourCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.Monitoring_NodeHour != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.Monitoring_NodeHour))
	}
	return &schema.CostComponent{
		Name:            "Node Hours",
		Unit:            "Node Hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("NODE_HOURS"),
		},
	}
}

func GetSysdigCostComponenets(r *ResourceInstance) []*schema.CostComponent {

	if r.Plan == "lite" {
		costComponent := &schema.CostComponent{
			Name:            "Lite plan",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		}
		costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
		return []*schema.CostComponent{costComponent}
	} else {
		return []*schema.CostComponent{
			GetSysdigTimeseriesCostComponent(r),
			GetSysdigContainerCostComponent(r),
			GetSysdigApiCallCostComponent(r),
			GetSysdigNodeHourCostComponent(r),
		}
	}
}
