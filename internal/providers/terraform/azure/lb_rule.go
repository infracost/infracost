package azure

import (
	"strings"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMLoadBalancerRuleRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_lb_rule",
		RFunc: NewAzureRMLoadBalancerRule,
		ReferenceAttributes: []string{
			"loadbalancer_id",
		},
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			return lookupRegion(d, []string{"loadbalancer_id"})
		},
	}
}

func NewAzureRMLoadBalancerRule(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Region
	region = convertRegion(region)

	lbSku := getParentLbSku(d.References("loadbalancer_id"))

	if lbSku == "" || strings.ToLower(lbSku) == "basic" {
		return &schema.Resource{
			Name:      d.Address,
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	var costComponents []*schema.CostComponent
	costComponents = append(costComponents, rulesCostComponent(region))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func getParentLbSku(lb []*schema.ResourceData) string {
	if len(lb) != 1 {
		return ""
	}

	if lb[0].Get("sku").Type != gjson.Null {
		return lb[0].Get("sku").String()
	}

	return "Basic" // default to basic
}

func rulesCostComponent(region string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Rule usage",
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Load Balancer"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr("Standard")},
				{Key: "meterName", ValueRegex: regexPtr("Overage LB Rules and Outbound Rules$")},
			},
		},
	}
}
