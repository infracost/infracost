package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func GetWafv2WebACLRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_wafv2_web_acl",
		RFunc: NewWafv2WebACL,
		Notes: []string{
			"Seller fees for Managed Rule Groups from AWS Marketplace are not included. Bot Control is not supported by Terraform.",
		},
	}
}

func NewWafv2WebACL(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	var costComponents []*schema.CostComponent
	var ruleGroupRules, managedRuleGroupRules, monthlyRequests, rule *decimal.Decimal
	var sumForRules decimal.Decimal

	costComponents = append(costComponents, wafv2WebACLCostComponent(
		region,
		"Web ACL usage",
		"months",
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
		costComponents = append(costComponents, wafv2WebACLUsageCostComponent(
			region,
			"Rules",
			"months",
			"USE1-RuleV2",
			1,
			&sumForRules,
		))
	}

	if d.Get("rule.0.statement.0.rule_group_reference_statement").Type != gjson.Null {
		counter := 0
		log.Warnf(">>>> processing resource=%s", d.Address)
		if d.Get("rule").Type != gjson.Null {
			rules := d.Get("rule").Array()
			for _, rule := range rules {
				log.Warnf(">>>> processing rule=%s", rule)
				if rule.Get("statement").Type != gjson.Null {
					statements := rule.Get("statement").Array()
					for _, statement := range statements {
						log.Warnf(">>>> processing statement=%s", statement)
						if statement.Get("rule_group_reference_statement").Type != gjson.Null {
							managedRuleGroupStatements := statement.Get("rule_group_reference_statement").Array()
							for _, managedRuleGroupStatement := range managedRuleGroupStatements {
								log.Warnf(">>>> FOUND a rule_group_reference_statement=%s", managedRuleGroupStatement.Get("statement.rule_group_reference_statement").String())
								counter++
							}
						}
					}
				}
			}
		}
		log.Warnf(">>>> TOTAL for RESOURCE=%s, rule_group_reference_statements=%d", d.Address, counter)

		if counter > 0 {
			costComponents = append(costComponents, wafv2WebACLCostComponent(
				region,
				"Rule groups",
				"months",
				"USE1-RuleV2",
				counter,
			))
		}
	}
	manageQuantity := d.Get("rule.0.statement.0.managed_rule_group_statement.0.name").Array()

	if len(manageQuantity) > 0 {
		costComponents = append(costComponents, wafv2WebACLCostComponent(
			region,
			"Managed rule groups",
			"months",
			"USE1-RuleV2",
			len(manageQuantity),
		))
	}

	if u != nil && u.Get("monthly_requests").Type != gjson.Null {
		monthlyRequests = decimalPtr(decimal.NewFromInt(u.Get("monthly_requests").Int()))
	}

	costComponents = append(costComponents, wafv2WebACLUsageCostComponent(
		region,
		"Requests",
		"1M requests",
		"USE1-RequestV2-Tier1",
		1000000,
		monthlyRequests,
	))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func wafv2WebACLUsageCostComponent(region, displayName, unit, usagetype string, unitMultiplier int, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            displayName,
		Unit:            unit,
		UnitMultiplier:  unitMultiplier,
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
func wafv2WebACLCostComponent(region, displayName, unit, usagetype string, quantity int) *schema.CostComponent {
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
