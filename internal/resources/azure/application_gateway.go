package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/usage"
)

type ApplicationGateway struct {
	Address                string
	SKUName                string
	SKUCapacity            int64
	AutoscalingMinCapacity *int64
	Region                 string
	MonthlyDataProcessedGB *float64 `infracost_usage:"monthly_data_processed_gb"`
	CapacityUnits          *int64   `infracost_usage:"capacity_units"`
}

func (r *ApplicationGateway) CoreType() string {
	return "ApplicationGateway"
}

func (r *ApplicationGateway) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{{Key: "monthly_data_processed_gb", ValueType: schema.Float64, DefaultValue: 0}, {Key: "monthly_v2_capacity_units", ValueType: schema.Int64, DefaultValue: 0}}
}

func (r *ApplicationGateway) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ApplicationGateway) BuildResource() *schema.Resource {
	var sku, tier string
	costComponents := make([]*schema.CostComponent, 0)

	skuNameParts := strings.Split(r.SKUName, "_")
	if len(skuNameParts) > 1 {
		sku = strings.ToLower(skuNameParts[1])
	}

	if strings.ToLower(skuNameParts[0]) == "standard" {
		tier = "basic"
	} else {
		tier = "WAF"
	}

	capacityUnits := int64(1)
	if r.SKUCapacity > 0 {
		capacityUnits = r.SKUCapacity
	} else if r.CapacityUnits != nil {
		capacityUnits = *r.CapacityUnits
	} else if r.AutoscalingMinCapacity != nil {
		capacityUnits = *r.AutoscalingMinCapacity
	}

	if sku == "v2" {
		costComponents = append(costComponents, r.v2CostComponents(skuNameParts, capacityUnits)...)
	} else {
		costComponents = append(costComponents, r.v1CostComponents(tier, sku, capacityUnits)...)
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *ApplicationGateway) v1CostComponents(tier, sku string, capacityUnits int64) []*schema.CostComponent {
	costComponents := make([]*schema.CostComponent, 0)
	var monthlyDataProcessedGb *decimal.Decimal
	tierLimits := []int{10240, 30720}

	costComponents = append(costComponents, r.gatewayCostComponent(fmt.Sprintf("Gateway usage (%s, %s)", tier, sku), tier, sku, capacityUnits))

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

	return costComponents
}

func (r *ApplicationGateway) v2CostComponents(skuNameParts []string, capacityUnits int64) []*schema.CostComponent {
	costComponents := make([]*schema.CostComponent, 0)

	var tier string
	if len(skuNameParts) > 0 && strings.ToLower(skuNameParts[0]) == "standard" {
		tier = "basic v2"
		costComponents = append(costComponents, r.fixedForV2CostComponent(fmt.Sprintf("Gateway usage (%s)", tier), "standard v2"))
		costComponents = append(costComponents, r.capacityUnitsCostComponent("basic", "standard v2", capacityUnits))
	} else {
		tier = "WAF v2"
		costComponents = append(costComponents, r.fixedForV2CostComponent(fmt.Sprintf("Gateway usage (%s)", tier), tier))
		costComponents = append(costComponents, r.capacityUnitsCostComponent("WAF", tier, capacityUnits))
	}
	return costComponents
}

func (r *ApplicationGateway) gatewayCostComponent(name, tier, sku string, capacityUnits int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(capacityUnits)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Application Gateway"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", ValueRegex: regexPtr(fmt.Sprintf("%s Application Gateway$", tier))},
				{Key: "meterName", ValueRegex: regexPtr(fmt.Sprintf("%s Gateway$", sku))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
func (r *ApplicationGateway) dataProcessingCostComponent(name, sku, startUsage string, qty *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Application Gateway"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", ValueRegex: regexPtr(fmt.Sprintf("%s Data Processed", sku))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr(startUsage),
		},
		UsageBased: true,
	}
}
func (r *ApplicationGateway) capacityUnitsCostComponent(name, tier string, capacityUnits int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           fmt.Sprintf("V2 capacity units (%s)", name),
		Unit:           "CU",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(capacityUnits)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Application Gateway"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", ValueRegex: regexPtr(fmt.Sprintf("Application Gateway %s$", tier))},
				{Key: "meterName", ValueRegex: regexPtr("Capacity Units$")},
			},
		},

		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
		UsageBased: true,
	}
}

func (r *ApplicationGateway) fixedForV2CostComponent(name, tier string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Application Gateway"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", ValueRegex: regexPtr(fmt.Sprintf("Application Gateway %s$", tier))},
				{Key: "meterName", ValueRegex: regexPtr("Fixed Cost$")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
