package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getCognitiveDeploymentRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_cognitive_deployment",
		CoreRFunc: newCognitiveDeployment,
		ReferenceAttributes: []string{
			"cognitive_account_id",
		},
	}
}

func newCognitiveDeployment(d *schema.ResourceData) schema.CoreResource {
	region := lookupRegion(d, []string{"cognitive_account_id"})

	cognitiveAccountRefs := d.References("cognitive_account_id")
	if region == "" && len(cognitiveAccountRefs) > 0 {
		region = lookupRegion(cognitiveAccountRefs[0], []string{"resource_group_name"})
	}

	return &azure.CognitiveDeployment{
		Address: d.Address,
		Region:  region,
		Model:   strings.ToLower(d.Get("model.0.name").String()),
		Version: strings.ToLower(d.Get("model.0.version").String()),
		Tier:    strings.ToLower(d.Get("scale.0.tier").String()),
	}
}
