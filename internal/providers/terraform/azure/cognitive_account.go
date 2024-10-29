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
	region := d.Region
	kind := d.Get("kind").String()

	if strings.EqualFold(kind, "speechservices") {
		return &azure.CognitiveAccountSpeech{
			Address: d.Address,
			Region:  region,
			Sku:     d.Get("sku_name").String(),
		}
	}

	if strings.EqualFold(kind, "luis") {
		return &azure.CognitiveAccountLUIS{
			Address: d.Address,
			Region:  region,
			Sku:     d.Get("sku_name").String(),
		}
	}

	if strings.EqualFold(kind, "textanalytics") {
		return &azure.CognitiveAccountLanguage{
			Address: d.Address,
			Region:  region,
			Sku:     d.Get("sku_name").String(),
		}
	}

	if strings.EqualFold(kind, "openai") {
		// OpenAI costs are counted as part of a Cognitive Deployment so
		// this resource is counted as free
		return schema.BlankCoreResource{
			Name: d.Address,
			Type: d.Type,
		}
	}

	logging.Logger.Warn().Msgf("Skipping resource %s. Kind %q is not supported", d.Address, kind)

	return nil
}
