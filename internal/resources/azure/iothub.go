package azure

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// IoTHub struct represents an IoT Hub
//
// Resource information: https://azure.microsoft.com/en-us/services/iot-hub/
// Pricing information: https://azure.microsoft.com/en-us/pricing/details/iot-hub/
type IoTHub struct {
	Address  string
	Region   string
	Sku      string
	Capacity int64
	DPS      bool

	Operations *int64 `infracost_usage:"monthly_operations"`
}

func (r *IoTHub) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

var OperationsUsageSchema = []*schema.UsageItem{
	{Key: "monthly_operations", DefaultValue: 0, ValueType: schema.Float64},
}

func (r *IoTHub) BuildResource() *schema.Resource {
	t := &schema.Resource{
		Name:           r.Address,
		UsageSchema:    OperationsUsageSchema,
		CostComponents: r.costComponents(),
	}

	if !r.DPS {
		schema.MultiplyQuantities(t, decimal.NewFromInt(r.Capacity))
	}

	return t
}

func (r *IoTHub) costComponents() []*schema.CostComponent {
	if r.DPS {
		return r.iotHubDPSCostComponent()
	}
	return r.iotHubCostComponent()
}

func (r *IoTHub) iotHubDPSCostComponent() []*schema.CostComponent {
	var quantity *decimal.Decimal
	itemsPerCost := 1000

	value := decimal.NewFromInt(*r.Operations)
	quantity = decimalPtr(value.Div(decimal.NewFromInt(int64(itemsPerCost))))

	costComponents := []*schema.CostComponent{
		{
			Name:            fmt.Sprintf("Instance (on-demand, %s)", r.Sku),
			Unit:            "1k operations",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: quantity,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(r.Region),
				Service:       strPtr("IoT Hub"),
				ProductFamily: strPtr("Internet of Things"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "skuName", Value: strPtr(r.Sku)},
					{Key: "meterName", ValueRegex: regexPtr("Operations$")},
				},
			},
		},
	}

	return costComponents
}

func (r *IoTHub) iotHubCostComponent() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{
		{
			Name:            fmt.Sprintf("Instance (on-demand, %s)", r.Sku),
			Unit:            "month",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(r.Region),
				Service:       strPtr("IoT Hub"),
				ProductFamily: strPtr("Internet of Things"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "skuName", Value: strPtr(r.Sku)},
					{Key: "meterName", ValueRegex: regexPtr("Unit$")},
				},
			},
		},
	}

	return costComponents
}
