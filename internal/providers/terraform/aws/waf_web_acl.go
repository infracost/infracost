package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetWafWebACLRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_waf_web_acl",
		RFunc: NewWafWebACL,
		Notes: []string{
			"Seller fees for Managed Rule Groups from AWS Marketplace are not included. Bot Control is not supported by Terraform.",
		},
	}
}

func NewWafWebACL(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	var costComponents []*schema.CostComponent
	var ruleGroupRules, monthlyRequests, rule *decimal.Decimal
	var ruleforGroup int

	costComponents = append(costComponents, wafWebACLCostComponent(
		region,
		"Web ACL usage",
		"hours",
		"USE1-WebACL",
		1,
	))
	var rules []gjson.Result
	if d.Get("rules").Type != gjson.Null {
		rules = d.Get("rules").Array()
	}
	if d.Get("rules.0.type").Type != gjson.Null {

		var count int
		for _, val := range rules {
			types := val.Get("type").String()

			if types == "REGULAR" || types == "RATE_BASED" {
				count++
			}

		}
		rule = decimalPtr(decimal.NewFromInt(int64(count)))
	}

	if u != nil && u.Get("rule_group_rules").Type != gjson.Null {
		ruleGroupRules = decimalPtr(decimal.NewFromInt(u.Get("rule_group_rules").Int()))
		sum := ruleGroupRules.Add(*rule)
		costComponents = append(costComponents, wafWebACLUsageCostComponent(
			region,
			"Rules",
			"hours",
			"USE1-Rule",
			&sum,
		))
	}

	if ruleGroupRules == nil {
		costComponents = append(costComponents, wafWebACLUsageCostComponent(
			region,
			"Rules",
			"hours",
			"USE1-Rule",
			rule,
		))
	}

	if d.Get("rules.0.type").Type != gjson.Null {
		rules := d.Get("rules").Array()
		var count int
		for _, val := range rules {
			types := val.Get("type").String()
			if types == "GROUP" {
				count++
			}
		}
		ruleforGroup = count
		if ruleforGroup > 0 {
			costComponents = append(costComponents, wafWebACLCostComponent(
				region,
				"Rule groups",
				"hours",
				"USE1-Rule",
				ruleforGroup,
			))
		}
	}

	if u != nil && u.Get("monthly_requests").Type != gjson.Null {
		monthlyRequests = decimalPtr(decimal.NewFromInt(u.Get("monthly_requests").Int()))
	}

	costComponents = append(costComponents, wafWebACLUsageCostComponent(
		region,
		"Requests",
		"requests",
		"USE1-Request",
		monthlyRequests,
	))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func wafWebACLUsageCostComponent(region, displayName, unit, usagetype string, quantity *decimal.Decimal) *schema.CostComponent {
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
func wafWebACLCostComponent(region, displayName, unit, usagetype string, quantity int) *schema.CostComponent {
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
