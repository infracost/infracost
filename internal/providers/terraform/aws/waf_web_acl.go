package aws

import (
	"fmt"
	"strings"

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

	costComponents = append(costComponents, wafWebACLUsageCostComponent(
		region,
		"Web ACL usage",
		"months",
		"USE1-WebACL",
		1,
		decimalPtr(decimal.NewFromInt(1)),
	))
	var rules []gjson.Result
	if d.Get("rules").Type != gjson.Null {
		rules = d.Get("rules").Array()
		var count int
		for _, val := range rules {
			types := val.Get("type").String()

			if strings.ToLower(types) == "regular" || strings.ToLower(types) == "rate_based" {
				count++
			}

		}
		rule = decimalPtr(decimal.NewFromInt(int64(count)))
	}

	if u != nil && u.Get("rule_group_rules").Type != gjson.Null {
		ruleGroupRules = decimalPtr(decimal.NewFromInt(u.Get("rule_group_rules").Int()))
		rule = decimalPtr(ruleGroupRules.Add(*rule))
	}

	costComponents = append(costComponents, wafWebACLUsageCostComponent(
		region,
		"Rules",
		"months",
		"USE1-Rule",
		1,
		rule,
	))

	rules = d.Get("rules").Array()
	var count int
	for _, val := range rules {
		types := val.Get("type").String()
		if strings.ToLower(types) == "group" {
			count++
		}
	}

	if count > 0 {
		costComponents = append(costComponents, wafWebACLUsageCostComponent(
			region,
			"Rule groups",
			"months",
			"USE1-Rule",
			1,
			decimalPtr(decimal.NewFromInt(int64(count))),
		))
	}

	if u != nil && u.Get("monthly_requests").Type != gjson.Null {
		monthlyRequests = decimalPtr(decimal.NewFromInt(u.Get("monthly_requests").Int()))
	}

	costComponents = append(costComponents, wafWebACLUsageCostComponent(
		region,
		"Requests",
		"1M requests",
		"USE1-Request",
		1000000,
		monthlyRequests,
	))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func wafWebACLUsageCostComponent(region, displayName, unit, usagetype string, unitMultiplier int, quantity *decimal.Decimal) *schema.CostComponent {
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
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", usagetype))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}
