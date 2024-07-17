package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func getAzureRMFirewallPolicyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "azurerm_firewall_policy",
		RFunc:               newAzureRMFirewallPolicy,
		ReferenceAttributes: []string{"azurerm_firewall_policy_rule_collection_group.firewall_policy_id"},
	}
}

func newAzureRMFirewallPolicy(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:         d.Address,
		ResourceType: d.Type,
		Tags:         d.Tags,
		DefaultTags:  d.DefaultTags,
		IsSkipped:    true,
		NoPrice:      true,
		SkipMessage:  "Free resource.",
	}
}
