package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getCognitiveAccountRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_cognitive_account",
		CoreRFunc: newCognitiveAccount,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newCognitiveAccount(d *schema.ResourceData) schema.CoreResource {
	region := lookupRegion(d, []string{"resource_group_name"})
	kind := d.Get("kind").String()

	if strings.EqualFold(kind, "speechservices") {
		return &azure.CognitiveAccountSpeech{
			Address: d.Address,
			Region:  region,
			Sku:     d.Get("sku_name").String(),
		}
	}

	logging.Logger.Warn().Msgf("Skipping resource %s. Kind '%s' is not supported", d.Address, kind)

	return nil
}
