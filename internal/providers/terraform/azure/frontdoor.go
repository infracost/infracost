package azure

import (
	"strings"

	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

// getFrontdoorRegistryItem returns a registry item for the resource
func getFrontdoorRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_frontdoor",
		CoreRFunc: newFrontdoor,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

// newFrontdoor parses Terraform's data and uses it to build a new resource
func newFrontdoor(d *schema.ResourceData) schema.CoreResource {
	region := d.Region

	if strings.HasPrefix(strings.ToLower(region), "usgov") {
		region = "US Gov Zone 1"
	} else {
		region = regionToCDNZone(region)
	}

	rulesCounter := 0
	rules := d.Get("routing_rule").Array()
	for _, rule := range rules {
		enabled := rule.Get("enabled").Type
		// if enabled is null this means the user has specified it and this resource is coming
		// from a hcl parsing. The default option is true, so increment the rulesCounter.
		if enabled == gjson.True || enabled == gjson.Null {
			rulesCounter++
		}
	}

	r := &azure.Frontdoor{
		Address:       d.Address,
		Region:        region,
		FrontendHosts: len(d.Get("frontend_endpoint").Array()),
		RoutingRules:  rulesCounter,
	}
	return r
}
