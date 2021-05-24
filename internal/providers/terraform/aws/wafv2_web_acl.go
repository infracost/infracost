package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetWafv2WevAclRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_wafv2_web_acl",
		RFunc: NewWafv2WebAcl,
		Notes: []string{
			"Seller fees for Managed Rule Groups from AWS Marketplace are not included. Bot Control is not supported by Terraform.",
		},
	}
}

func NewWafv2WebAcl(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	var costComponents []*schema.CostComponent
	var ruleGroupRules, managedRuleGroupRules, monthlyRequests, rule *decimal.Decimal
	var sumForRules decimal.Decimal

	costComponents = append(costComponents, wafv2WebAclCostComponent(
		region,
		"Web ACL usage",
		"hours",
		"USE1-WebACLV2",
		1,
	))
	if d.Get("rule.0.action").Type != gjson.Null {
		rules := d.Get("rule.0.action").Array()
		rule = decimalPtr(decimal.NewFromInt(int64(len(rules))))
	}
	if u != nil && u.Get("rule_group_rules").Type != gjson.Null && u.Get("managed_rule_group_rules").Type != gjson.Null {
		ruleGroupRules = decimalPtr(decimal.NewFromInt(u.Get("rule_group_rules").Int()))
		managedRuleGroupRules = decimalPtr(decimal.NewFromInt(u.Get("managed_rule_group_rules").Int()))
		sumForRules = ruleGroupRules.Add(*managedRuleGroupRules)
		if rule.IsPositive() {
			sumForRules = sumForRules.Add(*rule)
		}
	}

	if sumForRules.IsPositive() {
		costComponents = append(costComponents, wafv2WebAclUsageCostComponent(
			region,
			"Rules",
			"hours",
			"USE1-RuleV2",
			&sumForRules,
		))
	}

	var quantity int
	if d.Get("rule.0.statement.0.rule_group_reference_statement.0.arn").Type != gjson.Null {
		quantity += 1
	}

	if quantity > 0 {
		costComponents = append(costComponents, wafv2WebAclCostComponent(
			region,
			"Rule groups",
			"hours",
			"USE1-RuleV2",
			quantity,
		))
	}

	manageQuantity := d.Get("rule.0.statement.0.managed_rule_group_statement.0.name").Array()

	if len(manageQuantity) > 0 {
		costComponents = append(costComponents, wafv2WebAclCostComponent(
			region,
			"Managed rule groups",
			"hours",
			"USE1-RuleV2",
			len(manageQuantity),
		))
	}

	if u != nil && u.Get("monthly_requests").Type != gjson.Null {
		monthlyRequests = decimalPtr(decimal.NewFromInt(u.Get("monthly_requests").Int()))
	}

	costComponents = append(costComponents, wafv2WebAclUsageCostComponent(
		region,
		"Requests",
		"requests",
		"USE1-RequestV2-Tier1",
		monthlyRequests,
	))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func wafv2WebAclUsageCostComponent(region, displayName, unit, usagetype string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            displayName,
		Unit:            unit,
		UnitMultiplier:  1,
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("awswaf"),
			ProductFamily: strPtr("Web Application Firewall"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", Value: strPtr(usagetype)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}
func wafv2WebAclCostComponent(region, displayName, unit, usagetype string, quantity int) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            displayName,
		Unit:            unit,
		UnitMultiplier:  1,
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(int64(quantity))),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("awswaf"),
			ProductFamily: strPtr("Web Application Firewall"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", Value: strPtr(usagetype)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}
