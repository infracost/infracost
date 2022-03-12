package azure

import (
	"strings"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMApplicationInsightsWebRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_application_insights_web_test",
		RFunc: NewAzureRMApplicationInsightsWeb,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func NewAzureRMApplicationInsightsWeb(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{"resource_group_name"})
	costComponents := []*schema.CostComponent{}

	if d.Get("kind").Type != gjson.Null {
		if strings.ToLower(d.Get("kind").String()) == "multistep" && d.Get("enabled").Type == gjson.True {
			costComponents = append(costComponents, appInsightCostComponents(
				region,
				"Multi-step web test",
				"test",
				"Multi-step Web Test",
				"Enterprise",
				decimalPtr(decimal.NewFromInt(1))))
		}
	}

	if len(costComponents) == 0 {
		return &schema.Resource{
			Name:      d.Address,
			IsSkipped: true,
			NoPrice:   true,
		}
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}

}
