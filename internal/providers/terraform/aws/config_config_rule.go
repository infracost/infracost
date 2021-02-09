package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetConfigRuleItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_config_config_rule",
		RFunc: NewConfigRule,
	}
}

func NewConfigRule(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	var monthlyConfigItems *decimal.Decimal
	if u != nil && u.Get("monthly_config_items").Exists() {
		monthlyConfigItems = decimalPtr(decimal.NewFromInt(u.Get("monthly_config_items").Int()))
	}

	var monthlyCustomConfigItems *decimal.Decimal
	if u != nil && u.Get("monthly_custom_config_items").Exists() {
		monthlyCustomConfigItems = decimalPtr(decimal.NewFromInt(u.Get("monthly_custom_config_items").Int()))
	}

	monthlyConfigRules := decimal.Zero
	if u != nil && u.Get("monthly_rule_evaluations").Exists() {
		monthlyConfigRules = decimal.NewFromInt(u.Get("monthly_rule_evaluations").Int())
	}

	monthlyConformancePacks := decimal.Zero
	if u != nil && u.Get("monthly_conformance_pack_evaluations").Exists() {
		monthlyConformancePacks = decimal.NewFromInt(u.Get("monthly_conformance_pack_evaluations").Int())
	}

	var rulesTiers = map[string]decimal.Decimal{
		"tierOne":   decimal.Zero,
		"tierTwo":   decimal.Zero,
		"tierThree": decimal.Zero,
	}

	var packsTiers = map[string]decimal.Decimal{
		"tierOne":   decimal.Zero,
		"tierTwo":   decimal.Zero,
		"tierThree": decimal.Zero,
	}

	configRulesTiers := []int64{100000, 400000}
	conformancePacksTiers := []int64{1000000, 24000000}

	configRulesQuantities := calculateTierItems(monthlyConfigRules, rulesTiers, configRulesTiers)
	conformancePacksQuantities := calculateTierItems(monthlyConformancePacks, packsTiers, conformancePacksTiers)

	rulesTierOne := configRulesQuantities["tierOne"]
	rulesTierTwo := configRulesQuantities["tierTwo"]
	rulesTierThree := configRulesQuantities["tierThree"]

	packsTierOne := conformancePacksQuantities["tierOne"]
	packsTierTwo := conformancePacksQuantities["tierTwo"]
	packsTierThree := conformancePacksQuantities["tierThree"]

	costComponents := []*schema.CostComponent{}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Configuration items",
		Unit:            "records",
		UnitMultiplier:  1,
		MonthlyQuantity: monthlyConfigItems,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AWSConfig"),
			ProductFamily: strPtr("Management Tools - AWS Config"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/ConfigurationItemRecorded/")},
			},
		},
	})

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Custom configuration items",
		Unit:            "records",
		UnitMultiplier:  1,
		MonthlyQuantity: monthlyCustomConfigItems,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AWSConfig"),
			ProductFamily: strPtr("Management Tools - AWS Config"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/CustomConfigItemRecorded/")},
			},
		},
	})

	if rulesTierOne.GreaterThan(decimal.NewFromInt(0)) {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Config rule evaluations (first 100K)",
			Unit:            "evaluations",
			UnitMultiplier:  1,
			MonthlyQuantity: &rulesTierOne,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AWSConfig"),
				ProductFamily: strPtr("Management Tools - AWS Config Rules"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/ConfigRuleEvaluations/")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				StartUsageAmount: strPtr("0"),
				EndUsageAmount:   strPtr("100000"),
			},
		})
	}

	if rulesTierTwo.GreaterThan(decimal.NewFromInt(0)) {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Config rule evaluations (next 400K)",
			Unit:            "evaluations",
			UnitMultiplier:  1,
			MonthlyQuantity: &rulesTierTwo,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AWSConfig"),
				ProductFamily: strPtr("Management Tools - AWS Config Rules"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/ConfigRuleEvaluations/")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				StartUsageAmount: strPtr("100000"),
				EndUsageAmount:   strPtr("500000"),
			},
		})
	}

	if rulesTierThree.GreaterThan(decimal.NewFromInt(0)) {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Config rule evaluations (over 500K)",
			Unit:            "evaluations",
			UnitMultiplier:  1,
			MonthlyQuantity: &rulesTierThree,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AWSConfig"),
				ProductFamily: strPtr("Management Tools - AWS Config Rules"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/ConfigRuleEvaluations/")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				StartUsageAmount: strPtr("500000"),
			},
		})
	}

	if packsTierOne.GreaterThan(decimal.NewFromInt(0)) {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Conformance pack evaluations (first 1M)",
			Unit:            "evaluations",
			UnitMultiplier:  1,
			MonthlyQuantity: &packsTierOne,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AWSConfig"),
				ProductFamily: strPtr("Management Tools - AWS Config Conformance Packs"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/ConformancePackEvaluations/")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				StartUsageAmount: strPtr("0"),
				EndUsageAmount:   strPtr("1000000"),
			},
		})
	}

	if packsTierTwo.GreaterThan(decimal.NewFromInt(0)) {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Conformance pack evaluations (next 24M)",
			Unit:            "evaluations",
			UnitMultiplier:  1,
			MonthlyQuantity: &packsTierTwo,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AWSConfig"),
				ProductFamily: strPtr("Management Tools - AWS Config Conformance Packs"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/ConformancePackEvaluations/")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				StartUsageAmount: strPtr("1000000"),
				EndUsageAmount:   strPtr("25000000"),
			},
		})
	}

	if packsTierThree.GreaterThan(decimal.NewFromInt(0)) {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Conformance pack evaluations (over 25M)",
			Unit:            "evaluations",
			UnitMultiplier:  1,
			MonthlyQuantity: &packsTierThree,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AWSConfig"),
				ProductFamily: strPtr("Management Tools - AWS Config Conformance Packs"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/ConformancePackEvaluations/")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				StartUsageAmount: strPtr("25000000"),
			},
		})
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func calculateTierItems(items decimal.Decimal, tiers map[string]decimal.Decimal, tierLimits []int64) map[string]decimal.Decimal {
	rulesTierOneLimit := decimal.NewFromInt(tierLimits[0])
	rulesTierTwoLimit := decimal.NewFromInt(tierLimits[1])

	if items.GreaterThanOrEqual(rulesTierOneLimit) {
		tiers["tierOne"] = rulesTierOneLimit
	} else {
		tiers["tierOne"] = items
		return tiers
	}

	if items.GreaterThanOrEqual(rulesTierTwoLimit) {
		tiers["tierTwo"] = rulesTierTwoLimit
	} else {
		tiers["tierTwo"] = items.Sub(rulesTierOneLimit)
		return tiers
	}

	if items.GreaterThanOrEqual(rulesTierOneLimit.Add(rulesTierTwoLimit)) {
		tiers["tierThree"] = items.Sub(rulesTierOneLimit.Add(rulesTierTwoLimit))
		return tiers
	}
	return tiers
}
