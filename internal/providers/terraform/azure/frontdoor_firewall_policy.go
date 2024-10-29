package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

// getFrontdoorFirewallPolicyRegistryItem returns a registry item for the
// resource
func getFrontdoorFirewallPolicyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_frontdoor_firewall_policy",
		CoreRFunc: newFrontdoorFirewallPolicy,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

// newFrontdoorFirewallPolicy parses Terraform's data and uses it to build
// a new resource
func newFrontdoorFirewallPolicy(d *schema.ResourceData) schema.CoreResource {
	region := d.Region

	if strings.HasPrefix(strings.ToLower(region), "usgov") {
		region = "US Gov Zone 1"
	} else {
		region = regionToCDNZone(region)
	}

	customRules := 0
	if rules := d.Get("custom_rule"); rules.Exists() {
		customRules = len(rules.Array())
	}

	managedRulesets := 0
	if rules := d.Get("managed_rule"); rules.Exists() {
		managedRulesets = len(rules.Array())
	}

	r := &azure.FrontdoorFirewallPolicy{
		Address:         d.Address,
		Region:          region,
		CustomRules:     customRules,
		ManagedRulesets: managedRulesets,
	}
	return r
}
