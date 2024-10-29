package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMLoadBalancerOutboundRuleRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_lb_outbound_rule",
		RFunc: NewAzureRMLoadBalancerOutboundRule,
		ReferenceAttributes: []string{
			"loadbalancer_id",
			"resource_group_name",
		},
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			return lookupRegion(d, []string{"loadbalancer_id", "resource_group_name"})
		},
	}
}

func NewAzureRMLoadBalancerOutboundRule(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
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
