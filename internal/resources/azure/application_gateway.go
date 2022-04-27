package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
)

type ApplicationGateway struct {
	Address                string
	Region                 string
	SKUName                string
	SKUCapacity            int64
	MonthlyDataProcessedGB *float64 `infracost_usage:"monthly_data_processed_gb"`
	MonthlyV2CapacityUnits *int64   `infracost_usage:"monthly_v2_capacity_units"`
}

var ApplicationGatewayUsageSchema = []*schema.UsageItem{{Key: "monthly_data_processed_gb", ValueType: schema.Float64, DefaultValue: 0}, {Key: "monthly_v2_capacity_units", ValueType: schema.Int64, DefaultValue: 0}}

func (r *ApplicationGateway) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ApplicationGateway) BuildResource() *schema.Resource {
	var monthlyDataProcessedGb, monthlyCapacityUnits *decimal.Decimal
	var sku, tier string
	costComponents := make([]*schema.CostComponent, 0)
	tierLimits := []int{10240, 30720}

	skuNameParts := strings.Split(r.SKUName, "_")
	if len(skuNameParts[1]) != 0 {
		sku = strings.ToLower(skuNameParts[1])
	}
	if sku != "v2" {
		if strings.ToLower(skuNameParts[0]) == "standard" {
			tier = "basic"
		} else {
			tier = "WAF"
		}
		costComponents = append(costComponents, r.gatewayCostComponent(fmt.Sprintf("Gateway usage (%s, %s)", tier, sku), tier, sku))

		if r.MonthlyDataProcessedGB != nil {
			monthlyDataProcessedGb = decimalPtr(decimal.NewFromFloat(*r.MonthlyDataProcessedGB))
			result := usage.CalculateTierBuckets(*monthlyDataProcessedGb, tierLimits)

			if sku == "small" {
				if result[0].GreaterThan(decimal.Zero) {
					costComponents = append(costComponents, r.dataProcessingCostComponent("Data processing (0-10TB)", sku, "0", &result[0]))
				}
				if result[1].GreaterThan(decimal.Zero) {
					costComponents = append(costComponents, r.dataProcessingCostComponent("Data processing (10-40TB)", sku, "0", &result[1]))
				}
				if result[2].GreaterThan(decimal.Zero) {
					costComponents = append(costComponents, r.dataProcessingCostComponent("Data processing (over 40TB)", sku, "0", &result[2]))
				}
			}

			if sku == "medium" {
				if result[1].GreaterThan(decimal.Zero) {
					costComponents = append(costComponents, r.dataProcessingCostComponent("Data processing (10-40TB)", sku, "10240", &result[1]))
				}
				if result[2].GreaterThan(decimal.Zero) {
					costComponents = append(costComponents, r.dataProcessingCostComponent("Data processing (over 40TB)", sku, "10240", &result[2]))
				}
			}

			if sku == "large" {
				if result[2].GreaterThan(decimal.Zero) {
					costComponents = append(costComponents, r.dataProcessingCostComponent("Data processing (over 40TB)", sku, "40960", &result[2]))
				}
			}

		} else {
			var unknown *decimal.Decimal
			costComponents = append(costComponents, r.dataProcessingCostComponent("Data processing (0-10TB)", sku, "0", unknown))
		}
	}
	if r.MonthlyV2CapacityUnits != nil {
		monthlyCapacityUnits = decimalPtr(decimal.NewFromInt(*r.MonthlyV2CapacityUnits))
	}
	if sku == "v2" {
		if strings.ToLower(skuNameParts[0]) == "standard" {
			tier = "basic v2"
			costComponents = append(costComponents, r.fixedForV2CostComponent(fmt.Sprintf("Gateway usage (%s)", tier), "standard v2"))
			costComponents = append(costComponents, r.capacityUnitsCostComponent("basic", "standard v2", monthlyCapacityUnits))
		} else {
			tier = "WAF v2"
			costComponents = append(costComponents, r.fixedForV2CostComponent(fmt.Sprintf("Gateway usage (%s)", tier), tier))
			costComponents = append(costComponents, r.capacityUnitsCostComponent("WAF", tier, monthlyCapacityUnits))
		}

	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents, UsageSchema: ApplicationGatewayUsageSchema,
	}
}

func (r *ApplicationGateway) gatewayCostComponent(name, tier, sku string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(r.SKUCapacity)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Application Gateway"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", ValueRegex: strPtr(fmt.Sprintf("/%s Application Gateway/i", tier))},
				{Key: "meterName", ValueRegex: strPtr(fmt.Sprintf("/%s Gateway/i", sku))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func (r *ApplicationGateway) dataProcessingCostComponent(name, sku, startUsage string, capacity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: capacity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Application Gateway"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", ValueRegex: strPtr(fmt.Sprintf("/%s Data Processed/i", sku))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr(startUsage),
		},
	}
}

func (r *ApplicationGateway) capacityUnitsCostComponent(name, tier string, capacity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("V2 capacity units (%s)", name),
		Unit:            "CU",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: capacity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Application Gateway"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr("Capacity Units")},
				{Key: "productName", ValueRegex: strPtr(fmt.Sprintf("/Application Gateway %s/i", tier))},
			},
		},

		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func (r *ApplicationGateway) fixedForV2CostComponent(name, tier string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(r.SKUCapacity)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Application Gateway"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", ValueRegex: strPtr(fmt.Sprintf("/Application Gateway %s/i", tier))},
				{Key: "meterName", Value: strPtr("Fixed Cost")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
